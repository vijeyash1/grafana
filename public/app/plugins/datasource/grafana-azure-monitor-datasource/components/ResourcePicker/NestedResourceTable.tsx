import { cx } from '@emotion/css';
import React from 'react';

import { LoadingPlaceholder, useStyles2 } from '@grafana/ui';

import NestedRows from './NestedRows';
import getStyles from './styles';
import { ResourceRow, ResourceRowGroup, ResourceRowType } from './types';

// a nested resource table represents both the main table and the "selection" section below the tablle
interface NestedResourceTableProps {
  rows: ResourceRowGroup;
  selectedRows: ResourceRowGroup;
  noHeader?: boolean;
  requestNestedRows: (row: ResourceRow) => Promise<void>;
  onRowSelectedChange: (row: ResourceRow, selected: boolean) => void;
  selectableEntryTypes: ResourceRowType[];
  isLoading: boolean;
}

const NestedResourceTable: React.FC<NestedResourceTableProps> = ({
  rows,
  selectedRows,
  noHeader,
  requestNestedRows,
  onRowSelectedChange,
  selectableEntryTypes,
  isLoading,
}) => {
  const styles = useStyles2(getStyles);

  return (
    <>
      <table className={styles.table}>
        {!noHeader && (
          <thead>
            <tr className={cx(styles.row, styles.header)}>
              <td className={styles.cell}>Scope</td>
              <td className={styles.cell}>Type</td>
              <td className={styles.cell}>Location</td>
            </tr>
          </thead>
        )}
      </table>

      <div className={styles.tableScroller}>
        <table className={styles.table}>
          <tbody>
            {isLoading && (
              <tr className={cx(styles.row)}>
                <td className={styles.cell}>
                  <LoadingPlaceholder text={'Loading...'} />
                </td>
              </tr>
            )}
            {!isLoading && rows.length === 0 && (
              <tr className={cx(styles.row)}>
                <td className={styles.cell}>No resources found</td>
              </tr>
            )}
            {!isLoading && rows.length > 0 && (
              <NestedRows
                rows={rows}
                selectedRows={selectedRows}
                level={0}
                requestNestedRows={requestNestedRows}
                onRowSelectedChange={onRowSelectedChange}
                selectableEntryTypes={selectableEntryTypes}
                scrollIntoView={!noHeader}
              />
            )}
          </tbody>
        </table>
      </div>
    </>
  );
};

export default NestedResourceTable;
