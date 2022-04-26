import React from 'react';

import NestedRow from './NestedRow';
import { ResourceRow, ResourceRowGroup, ResourceRowType } from './types';

interface NestedRowsProps {
  rows: ResourceRowGroup;
  level: number;
  selectedRows: ResourceRowGroup;
  requestNestedRows: (row: ResourceRow) => Promise<void>;
  onRowSelectedChange: (row: ResourceRow, selected: boolean) => void;
  selectableEntryTypes: ResourceRowType[];
  scrollIntoView: boolean;
}

const NestedRows: React.FC<NestedRowsProps> = ({
  rows,
  selectedRows,
  level,
  requestNestedRows,
  onRowSelectedChange,
  selectableEntryTypes,
  scrollIntoView,
}) => (
  <>
    {rows.map((row) => (
      <NestedRow
        key={row.uri}
        row={row}
        selectedRows={selectedRows}
        level={level}
        requestNestedRows={requestNestedRows}
        onRowSelectedChange={onRowSelectedChange}
        selectableEntryTypes={selectableEntryTypes}
        scrollIntoView={scrollIntoView}
      />
    ))}
  </>
);

export default NestedRows;
