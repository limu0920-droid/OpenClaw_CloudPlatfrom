export function formatAccessEntryType(entryType: string) {
  switch (entryType) {
    case 'web':
      return '网页版'
    case 'h5':
      return 'H5 入口'
    case 'admin':
      return '管理后台'
    case 'portal':
      return '用户入口'
    default:
      return entryType
  }
}
