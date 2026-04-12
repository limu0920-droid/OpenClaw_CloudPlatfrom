import type { InstanceAccess } from './types'

const workspacePriority = ['web', 'h5', 'portal']

export function getWorkspaceAccess(accesses: InstanceAccess[]) {
  for (const type of workspacePriority) {
    const match = accesses.find((item) => item.entryType === type)
    if (match) {
      return match
    }
  }

  return accesses.find((item) => item.isPrimary) ?? accesses[0] ?? null
}

export function getAdminAccess(accesses: InstanceAccess[]) {
  return accesses.find((item) => item.entryType === 'admin') ?? null
}
