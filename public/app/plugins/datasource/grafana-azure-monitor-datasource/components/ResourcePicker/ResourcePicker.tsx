import { css } from '@emotion/css';
import React, { useCallback, useEffect, useMemo, useState } from 'react';

import { GrafanaTheme2 } from '@grafana/data';
import { Alert, Button, Icon, Input, Tooltip, useStyles2, Collapse, Label } from '@grafana/ui';

import ResourcePickerData from '../../resourcePicker/resourcePickerData';
import messageFromError from '../../utils/messageFromError';
import { Space } from '../Space';

import NestedResourceTable from './NestedResourceTable';
import Search from './Search';
import { ResourceRow, ResourceRowGroup, ResourceRowType } from './types';
import { addResources, findRow } from './utils';

interface ResourcePickerProps {
  resourcePickerData: ResourcePickerData;
  resourceURI: string | undefined;
  selectableEntryTypes: ResourceRowType[];

  onApply: (resourceURI: string | undefined) => void;
  onCancel: () => void;
}

const ResourcePicker = ({
  resourcePickerData,
  resourceURI,
  onApply,
  onCancel,
  selectableEntryTypes,
}: ResourcePickerProps) => {
  const styles = useStyles2(getStyles);

  type LoadingStatus = 'NotStarted' | 'Started' | 'Done';
  const [loadingStatus, setLoadingStatus] = useState<LoadingStatus>('NotStarted');
  const [rows, setRows] = useState<ResourceRowGroup>([]);
  const [internalSelectedURI, setInternalSelectedURI] = useState<string | undefined>(resourceURI);
  const [errorMessage, setErrorMessage] = useState<string | undefined>(undefined);
  const [isAdvancedOpen, setIsAdvancedOpen] = useState(resourceURI?.includes('$'));

  // Sync the resourceURI prop to internal state
  useEffect(() => {
    setInternalSelectedURI(resourceURI);
  }, [resourceURI]);

  // Request initial data on first mount
  useEffect(() => {
    if (loadingStatus === 'NotStarted') {
      const loadInitialData = async () => {
        try {
          setLoadingStatus('Started');
          const resources = await resourcePickerData.fetchInitialRows(internalSelectedURI || '');
          setRows(resources);
          setLoadingStatus('Done');
        } catch (error) {
          setLoadingStatus('Done');
          setErrorMessage(messageFromError(error));
        }
      };

      loadInitialData();
    }
  }, [resourcePickerData, internalSelectedURI, loadingStatus]);

  // Map the selected item into an array of rows
  const selectedResourceRows = useMemo(() => {
    const found = internalSelectedURI && findRow(rows, internalSelectedURI);
    return found
      ? [
          {
            ...found,
            children: undefined,
          },
        ]
      : [];
  }, [internalSelectedURI, rows]);

  // Request resources for an expanded resource group
  const requestNestedRows = useCallback(
    async (expandedRow: ResourceRow) => {
      // clear error message (also when loading cached resources)
      setErrorMessage(undefined);

      // If we already have children, we don't need to re-fetch them.
      if (expandedRow.children?.length) {
        return;
      }

      try {
        const nestedRows = await resourcePickerData.fetchNestedRowData(expandedRow);
        const newRows = addResources(rows, expandedRow.uri, nestedRows);
        setRows(newRows);
      } catch (error) {
        setErrorMessage(messageFromError(error));
        throw error;
      }
    },
    [resourcePickerData, rows]
  );

  const handleSelectionChanged = useCallback((row: ResourceRow, isSelected: boolean) => {
    isSelected ? setInternalSelectedURI(row.uri) : setInternalSelectedURI(undefined);
  }, []);

  const handleApply = useCallback(() => {
    onApply(internalSelectedURI);
  }, [internalSelectedURI, onApply]);

  const handleSearch = useCallback(
    async (searchWord: string) => {
      if (!searchWord) {
        setLoadingStatus('NotStarted');
        return;
      }
      try {
        setLoadingStatus('Started');
        const searchResults = await resourcePickerData.search(searchWord, selectableEntryTypes);
        setRows(searchResults);
      } catch (err) {
        setErrorMessage(messageFromError(err));
      }
      setLoadingStatus('Done');
    },
    [resourcePickerData, setRows, setErrorMessage, setLoadingStatus, selectableEntryTypes]
  );

  return (
    <div>
      <>
        <Search searchFn={handleSearch} />
        <Space v={2} />

        <NestedResourceTable
          rows={rows}
          requestNestedRows={requestNestedRows}
          onRowSelectedChange={handleSelectionChanged}
          selectedRows={selectedResourceRows}
          selectableEntryTypes={selectableEntryTypes}
          isLoading={loadingStatus === 'Started'}
        />

        <div className={styles.selectionFooter}>
          {selectedResourceRows.length > 0 && (
            <>
              <h5>Selection</h5>
              <NestedResourceTable
                rows={selectedResourceRows}
                requestNestedRows={requestNestedRows}
                onRowSelectedChange={handleSelectionChanged}
                selectedRows={selectedResourceRows}
                noHeader={true}
                selectableEntryTypes={selectableEntryTypes}
                isLoading={false}
              />
              <Space v={2} />
            </>
          )}
          <Collapse
            collapsible
            label="Advanced"
            isOpen={isAdvancedOpen}
            onToggle={() => setIsAdvancedOpen(!isAdvancedOpen)}
          >
            <Label htmlFor={`input-${internalSelectedURI}`}>
              <h6>
                Resource URI{' '}
                <Tooltip
                  content={
                    <>
                      Manually edit the{' '}
                      <a
                        href="https://docs.microsoft.com/en-us/azure/azure-monitor/logs/log-standard-columns#_resourceid"
                        rel="noopener noreferrer"
                        target="_blank"
                      >
                        resource uri.{' '}
                      </a>
                      Supports the use of multiple template variables (ex: /subscriptions/$subId/resourceGroups/$rg)
                    </>
                  }
                  placement="right"
                  interactive={true}
                >
                  <Icon name="info-circle" />
                </Tooltip>
              </h6>
            </Label>
            <Input
              id={`input-${internalSelectedURI}`}
              value={internalSelectedURI}
              onChange={(event) => setInternalSelectedURI(event.currentTarget.value)}
              placeholder="ex: /subscriptions/$subId"
            />
          </Collapse>
          <Space v={2} />

          <Button disabled={!!errorMessage} onClick={handleApply}>
            Apply
          </Button>

          <Space layout="inline" h={1} />

          <Button onClick={onCancel} variant="secondary">
            Cancel
          </Button>
        </div>
      </>
      {errorMessage && (
        <>
          <Space v={2} />
          <Alert severity="error" title="An error occurred while requesting resources from Azure Monitor">
            {errorMessage}
          </Alert>
        </>
      )}
    </div>
  );
};

export default ResourcePicker;

const getStyles = (theme: GrafanaTheme2) => ({
  selectionFooter: css({
    position: 'sticky',
    bottom: 0,
    background: theme.colors.background.primary,
    paddingTop: theme.spacing(2),
  }),
  loadingWrapper: css({
    textAlign: 'center',
    paddingTop: theme.spacing(2),
    paddingBottom: theme.spacing(2),
    color: theme.colors.text.secondary,
  }),
});
