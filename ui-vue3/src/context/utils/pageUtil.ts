import context from '@/context'
export const setPageRowData = (key: string, value: unknown) => {
  if (!key) {
    return
  }
  if (!context.page.rowData) {
    context.page.rowData = {}
  }
  context.page.rowData[key] = value
}

export const setPageInfo = (key: string, value: unknown) => {
  if (!key) {
    return
  }
  if (!context.page.info) {
    context.page.info = {}
  }
  context.page.info[key] = value
}

export const homeOverviewHandle = (data: Record<string, unknown>) => {
  const res = `
  appications count: ${data.appCount}
  services count: ${data.serviceCount}
  Instances count: ${data.instanceCount}
  protocols distribution: ${JSON.stringify(data.protocols || [])}
  releases distribution: ${JSON.stringify(data.releases || [])}
  discoveries distribution: ${JSON.stringify(data.discoveries || [])}
  `
  return res
}
