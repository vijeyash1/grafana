package cloudwatch

import (
	"context"
	"encoding/json"
	"fmt"
	"testing"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/cloudwatch"
	"github.com/aws/aws-sdk-go/service/cloudwatch/cloudwatchiface"
	"github.com/grafana/grafana-plugin-sdk-go/backend"
	"github.com/grafana/grafana-plugin-sdk-go/backend/datasource"
	"github.com/grafana/grafana-plugin-sdk-go/backend/instancemgmt"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestTimeSeriesQuery(t *testing.T) {
	executor := newExecutor(nil, newTestConfig(), &fakeSessionCache{})
	now := time.Now()

	origNewCWClient := NewCWClient
	t.Cleanup(func() {
		NewCWClient = origNewCWClient
	})

	var cwClient fakeCWClient

	NewCWClient = func(sess *session.Session) cloudwatchiface.CloudWatchAPI {
		return &cwClient
	}

	t.Run("Custom metrics", func(t *testing.T) {
		cwClient = fakeCWClient{
			CloudWatchAPI: nil,
			GetMetricDataOutput: cloudwatch.GetMetricDataOutput{
				NextToken: nil,
				Messages:  []*cloudwatch.MessageData{},
				MetricDataResults: []*cloudwatch.MetricDataResult{
					{
						StatusCode: aws.String("Complete"), Id: aws.String("a"), Label: aws.String("NetworkOut"), Values: []*float64{aws.Float64(1.0)}, Timestamps: []*time.Time{&now},
					},
					{
						StatusCode: aws.String("Complete"), Id: aws.String("b"), Label: aws.String("NetworkIn"), Values: []*float64{aws.Float64(1.0)}, Timestamps: []*time.Time{&now},
					},
				},
			},
		}

		im := datasource.NewInstanceManager(func(s backend.DataSourceInstanceSettings) (instancemgmt.Instance, error) {
			return datasourceInfo{}, nil
		})

		executor := newExecutor(im, newTestConfig(), &fakeSessionCache{})
		resp, err := executor.QueryData(context.Background(), &backend.QueryDataRequest{
			PluginContext: backend.PluginContext{
				DataSourceInstanceSettings: &backend.DataSourceInstanceSettings{},
			},
			Queries: []backend.DataQuery{
				{
					RefID: "A",
					TimeRange: backend.TimeRange{
						From: now.Add(time.Hour * -2),
						To:   now.Add(time.Hour * -1),
					},
					JSON: json.RawMessage(`{
						"type":      "timeSeriesQuery",
						"subtype":   "metrics",
						"namespace": "AWS/EC2",
						"metricName": "NetworkOut",
						"expression": "",
						"dimensions": {
						  "InstanceId": "i-00645d91ed77d87ac"
						},
						"region": "us-east-2",
						"id": "a",
						"alias": "NetworkOut",
						"statistics": [
						  "Maximum"
						],
						"period": "300",
						"hide": false,
						"matchExact": true,
						"refId": "A"
					}`),
				},
				{
					RefID: "B",
					TimeRange: backend.TimeRange{
						From: now.Add(time.Hour * -2),
						To:   now.Add(time.Hour * -1),
					},
					JSON: json.RawMessage(`{
						"type":      "timeSeriesQuery",
						"subtype":   "metrics",
						"namespace": "AWS/EC2",
						"metricName": "NetworkIn",
						"expression": "",
						"dimensions": {
						"InstanceId": "i-00645d91ed77d87ac"
						},
						"region": "us-east-2",
						"id": "b",
						"alias": "NetworkIn",
						"statistics": [
						"Maximum"
						],
						"period": "300",
						"matchExact": true,
						"refId": "B"
					}`),
				},
			},
		})
		require.NoError(t, err)
		assert.Equal(t, "NetworkOut", resp.Responses["A"].Frames[0].Name)
		assert.Equal(t, "NetworkIn", resp.Responses["B"].Frames[0].Name)
	})

	t.Run("End time before start time should result in error", func(t *testing.T) {
		_, err := executor.executeTimeSeriesQuery(context.Background(), &backend.QueryDataRequest{Queries: []backend.DataQuery{{TimeRange: backend.TimeRange{
			From: now.Add(time.Hour * -1),
			To:   now.Add(time.Hour * -2),
		}}}})
		assert.EqualError(t, err, "invalid time range: start time must be before end time")
	})

	t.Run("End time equals start time should result in error", func(t *testing.T) {
		_, err := executor.executeTimeSeriesQuery(context.Background(), &backend.QueryDataRequest{Queries: []backend.DataQuery{{TimeRange: backend.TimeRange{
			From: now.Add(time.Hour * -1),
			To:   now.Add(time.Hour * -1),
		}}}})
		assert.EqualError(t, err, "invalid time range: start time must be before end time")
	})
}

func Test_QueryData_timeSeriesQuery_GetMetricDataWithContext_passes_query_alias_as_label(t *testing.T) {
	origNewCWClient := NewCWClient
	t.Cleanup(func() {
		NewCWClient = origNewCWClient
	})
	var cwClient fakeCWClient
	NewCWClient = func(sess *session.Session) cloudwatchiface.CloudWatchAPI {
		return &cwClient
	}

	testCases := map[string]string{
		"not-yet-migrated legacy alias": "{{ period  }} some words {{   InstanceId }}",
		"migrated dynamic labels alias": "${PROP('Period')} some words ${PROP('Dim.InstanceId')}",
	}
	for name, inputAlias := range testCases {
		t.Run(name, func(t *testing.T) {
			cwClient = fakeCWClient{}
			im := datasource.NewInstanceManager(func(s backend.DataSourceInstanceSettings) (instancemgmt.Instance, error) {
				return datasourceInfo{}, nil
			})
			executor := newExecutor(im, newTestConfig(), &fakeSessionCache{})

			_, err := executor.QueryData(context.Background(), &backend.QueryDataRequest{
				PluginContext: backend.PluginContext{DataSourceInstanceSettings: &backend.DataSourceInstanceSettings{}},
				Queries: []backend.DataQuery{
					{
						RefID: "A",
						TimeRange: backend.TimeRange{
							From: time.Now().Add(time.Hour * -2),
							To:   time.Now().Add(time.Hour * -1),
						},
						JSON: json.RawMessage(fmt.Sprintf(`{
						"type":      "timeSeriesQuery",
						"subtype":   "metrics",
						"namespace": "AWS/EC2",
						"metricName": "NetworkOut",
						"expression": "",
						"dimensions": {
						  "InstanceId": "i-00645d91ed77d87ac"
						},
						"region": "us-east-2",
						"id": "a",
						"alias": "%s",
						"statistics": [
						  "Maximum"
						],
						"period": "300",
						"hide": false,
						"matchExact": true,
						"refId": "A"
					}`, inputAlias)),
					},
				},
			})

			assert.NoError(t, err)
			assert.Len(t, cwClient.calls.getMetricDataWithContext, 1)
			assert.Len(t, cwClient.calls.getMetricDataWithContext[0].MetricDataQueries, 1)
			require.NotNil(t, cwClient.calls.getMetricDataWithContext[0].MetricDataQueries[0].Label)

			assert.Equal(t, "${PROP('Period')} some words ${PROP('Dim.InstanceId')}", *cwClient.calls.getMetricDataWithContext[0].MetricDataQueries[0].Label)
		})
	}
}

type queryDimensions struct {
	InstanceID []string `json:"InstanceId,omitempty"`
}

type queryParameters struct {
	MetricQueryType  metricQueryType  `json:"metricQueryType"`
	MetricEditorMode metricEditorMode `json:"metricEditorMode"`
	Dimensions       queryDimensions  `json:"dimensions"`
	Expression       string           `json:"expression"`
	Alias            string           `json:"alias"`
	Statistic        string           `json:"statistic"`
	Period           string           `json:"period"`
	MatchExact       bool             `json:"matchExact"`
	MetricName       string           `json:"metricName"`
}

var queryId = "query id"

func newTestQuery(t testing.TB, p queryParameters) json.RawMessage {
	t.Helper()

	tsq := struct {
		Type             string           `json:"type"`
		MetricQueryType  metricQueryType  `json:"metricQueryType"`
		MetricEditorMode metricEditorMode `json:"metricEditorMode"`
		Namespace        string           `json:"namespace"`
		MetricName       string           `json:"metricName"`
		Dimensions       struct {
			InstanceID []string `json:"InstanceId,omitempty"`
		} `json:"dimensions"`
		Expression string `json:"expression"`
		Region     string `json:"region"`
		ID         string `json:"id"`
		Alias      string `json:"alias"`
		Statistic  string `json:"statistic"`
		Period     string `json:"period"`
		MatchExact bool   `json:"matchExact"`
		RefID      string `json:"refId"`
	}{
		Type:   "timeSeriesQuery",
		Region: "us-east-2",
		ID:     queryId,
		RefID:  "A",

		MatchExact:       p.MatchExact,
		MetricQueryType:  p.MetricQueryType,
		MetricEditorMode: p.MetricEditorMode,
		Dimensions:       p.Dimensions,
		Expression:       p.Expression,
		Alias:            p.Alias,
		Statistic:        p.Statistic,
		Period:           p.Period,
		MetricName:       p.MetricName,
	}

	marshalled, err := json.Marshal(tsq)
	require.NoError(t, err)

	return marshalled
}

func Test_QueryData_response_data_frame_names(t *testing.T) {
	origNewCWClient := NewCWClient
	t.Cleanup(func() {
		NewCWClient = origNewCWClient
	})
	var cwClient fakeCWClient
	NewCWClient = func(sess *session.Session) cloudwatchiface.CloudWatchAPI {
		return &cwClient
	}
	labelFromGetMetricData := "some label"
	cwClient = fakeCWClient{
		GetMetricDataOutput: cloudwatch.GetMetricDataOutput{
			MetricDataResults: []*cloudwatch.MetricDataResult{
				{StatusCode: aws.String("Complete"), Id: aws.String(queryId), Label: aws.String(labelFromGetMetricData),
					Values: []*float64{aws.Float64(1.0)}, Timestamps: []*time.Time{{}}},
			},
		},
	}
	im := datasource.NewInstanceManager(func(s backend.DataSourceInstanceSettings) (instancemgmt.Instance, error) {
		return datasourceInfo{}, nil
	})
	executor := newExecutor(im, newTestConfig(), &fakeSessionCache{})

	t.Run("where no alias is provided and query is math expression, then frame name is queryId", func(t *testing.T) {
		query := newTestQuery(t, queryParameters{
			MetricQueryType:  MetricQueryTypeSearch,
			MetricEditorMode: MetricEditorModeRaw,
		})

		resp, err := executor.QueryData(context.Background(), &backend.QueryDataRequest{
			PluginContext: backend.PluginContext{DataSourceInstanceSettings: &backend.DataSourceInstanceSettings{}},
			Queries: []backend.DataQuery{
				{
					RefID:     "A",
					TimeRange: backend.TimeRange{From: time.Now().Add(time.Hour * -2), To: time.Now().Add(time.Hour * -1)},
					JSON:      query,
				},
			},
		})

		assert.NoError(t, err)
		assert.Equal(t, queryId, resp.Responses["A"].Frames[0].Name)
	})

	t.Run("where no alias provided and query type is MetricQueryTypeQuery, then frame name is label", func(t *testing.T) {
		query := newTestQuery(t, queryParameters{
			MetricQueryType: MetricQueryTypeQuery,
		})

		resp, err := executor.QueryData(context.Background(), &backend.QueryDataRequest{
			PluginContext: backend.PluginContext{DataSourceInstanceSettings: &backend.DataSourceInstanceSettings{}},
			Queries: []backend.DataQuery{
				{
					RefID:     "A",
					TimeRange: backend.TimeRange{From: time.Now().Add(time.Hour * -2), To: time.Now().Add(time.Hour * -1)},
					JSON:      query,
				},
			},
		})

		assert.NoError(t, err)
		assert.Equal(t, labelFromGetMetricData, resp.Responses["A"].Frames[0].Name)
	})

	// where query is inferred search expression and not multivalued dimension expression, then frame name is label
	testCasesReturningLabel := map[string]queryParameters{
		"with specific dimensions, matchExact false": {Dimensions: queryDimensions{[]string{"some-instance"}}, MatchExact: false},
		"with wildcard dimensions, matchExact false": {Dimensions: queryDimensions{[]string{"*"}}, MatchExact: false},
		"with wildcard dimensions, matchExact true":  {Dimensions: queryDimensions{[]string{"*"}}, MatchExact: true},
		"no dimension, matchExact false":             {Dimensions: queryDimensions{}, MatchExact: false},
	}
	for name, parameters := range testCasesReturningLabel {
		t.Run(name, func(t *testing.T) {
			query := newTestQuery(t, parameters)

			resp, err := executor.QueryData(context.Background(), &backend.QueryDataRequest{
				PluginContext: backend.PluginContext{DataSourceInstanceSettings: &backend.DataSourceInstanceSettings{}},
				Queries: []backend.DataQuery{
					{
						RefID:     "A",
						TimeRange: backend.TimeRange{From: time.Now().Add(time.Hour * -2), To: time.Now().Add(time.Hour * -1)},
						JSON:      query,
					},
				},
			})

			assert.NoError(t, err)
			assert.Equal(t, labelFromGetMetricData, resp.Responses["A"].Frames[0].Name)
		})
	}

	// complementary test cases to above return default of "metricName_stat"
	testCasesReturningMetricStat := map[string]queryParameters{
		"with specific dimensions, matchExact true": {
			Dimensions: queryDimensions{[]string{"some-instance"}},
			MatchExact: true,
			MetricName: "CPUUtilization",
			Statistic:  "Maximum",
		},
		"no dimensions, matchExact true": {
			Dimensions: queryDimensions{},
			MatchExact: true,
			MetricName: "CPUUtilization",
			Statistic:  "Maximum",
		},
		"multivalued dimensions, matchExact true": {
			Dimensions: queryDimensions{[]string{"some-instance", "another-instance"}},
			MatchExact: true,
			MetricName: "CPUUtilization",
			Statistic:  "Maximum",
		},
		"multivalued dimensions, matchExact false": {
			Dimensions: queryDimensions{[]string{"some-instance", "another-instance"}},
			MatchExact: false,
			MetricName: "CPUUtilization",
			Statistic:  "Maximum",
		},
	}
	for name, parameters := range testCasesReturningMetricStat {
		t.Run(name, func(t *testing.T) {
			query := newTestQuery(t, parameters)

			resp, err := executor.QueryData(context.Background(), &backend.QueryDataRequest{
				PluginContext: backend.PluginContext{DataSourceInstanceSettings: &backend.DataSourceInstanceSettings{}},
				Queries: []backend.DataQuery{
					{
						RefID:     "A",
						TimeRange: backend.TimeRange{From: time.Now().Add(time.Hour * -2), To: time.Now().Add(time.Hour * -1)},
						JSON:      query,
					},
				},
			})

			assert.NoError(t, err)
			assert.Equal(t, "CPUUtilization_Maximum", resp.Responses["A"].Frames[0].Name)
		})
	}
}
