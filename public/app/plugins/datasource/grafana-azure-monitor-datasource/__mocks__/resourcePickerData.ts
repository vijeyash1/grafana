import ResourcePickerData from '../resourcePicker/resourcePickerData';

import { createMockInstanceSetttings } from './instanceSettings';

type DeepPartial<T> = {
  [P in keyof T]?: DeepPartial<T[P]>;
};

export default function createMockResourcePickerData(overrides?: DeepPartial<ResourcePickerData>) {
  const rpd = new ResourcePickerData(createMockInstanceSetttings());
  const _mockResourcePicker: DeepPartial<ResourcePickerData> = {
    fetchInitialRows: rpd.fetchInitialRows,
    fetchNestedRowData: rpd.fetchNestedRowData,
    search: jest.fn().mockResolvedValue([]),
    getSubscriptions: jest.fn().mockResolvedValue([]),
    getResourceGroupsBySubscriptionId: jest.fn().mockResolvedValue([]),
    getResourcesForResourceGroup: jest.fn().mockResolvedValue([]),
    getResourceURIFromWorkspace: jest.fn().mockReturnValue(''),
    getResourceURIDisplayProperties: jest.fn().mockResolvedValue({}),
    ...overrides,
  };

  const mockDatasource = _mockResourcePicker as ResourcePickerData;

  return jest.mocked(mockDatasource, true);
}
