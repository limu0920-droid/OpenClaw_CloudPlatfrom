import type { WorkspaceScope, WorkspaceSession } from './types'

export interface WorkspaceSessionPreferenceState {
  pinnedIds: string[]
  favoriteIds: string[]
}

const workspacePreferenceStoragePrefix = 'openclaw.workspace.session-preferences'

function workspacePreferenceStorageKey(scope: WorkspaceScope) {
  return `${workspacePreferenceStoragePrefix}.${scope}`
}

function normalizeIdList(value: unknown) {
  if (!Array.isArray(value)) {
    return []
  }
  return value
    .map((item) => (typeof item === 'string' ? item.trim() : ''))
    .filter((item, index, items) => item && items.indexOf(item) === index)
}

function readWorkspacePreferenceState(scope: WorkspaceScope): WorkspaceSessionPreferenceState {
  if (typeof window === 'undefined') {
    return { pinnedIds: [], favoriteIds: [] }
  }

  try {
    const raw = window.localStorage.getItem(workspacePreferenceStorageKey(scope))
    if (!raw) {
      return { pinnedIds: [], favoriteIds: [] }
    }

    const parsed = JSON.parse(raw) as { pinnedIds?: unknown; favoriteIds?: unknown }
    return {
      pinnedIds: normalizeIdList(parsed.pinnedIds),
      favoriteIds: normalizeIdList(parsed.favoriteIds),
    }
  } catch {
    return { pinnedIds: [], favoriteIds: [] }
  }
}

function writeWorkspacePreferenceState(scope: WorkspaceScope, state: WorkspaceSessionPreferenceState) {
  if (typeof window === 'undefined') {
    return state
  }

  const normalized = {
    pinnedIds: normalizeIdList(state.pinnedIds),
    favoriteIds: normalizeIdList(state.favoriteIds),
  }

  try {
    window.localStorage.setItem(workspacePreferenceStorageKey(scope), JSON.stringify(normalized))
  } catch {
    // ignore storage failures
  }

  return normalized
}

function toggleId(ids: string[], id: string) {
  return ids.includes(id) ? ids.filter((item) => item !== id) : [id, ...ids]
}

export function loadWorkspaceSessionPreferences(scope: WorkspaceScope) {
  return readWorkspacePreferenceState(scope)
}

export function toggleWorkspaceSessionPinned(scope: WorkspaceScope, sessionId: string) {
  const current = readWorkspacePreferenceState(scope)
  return writeWorkspacePreferenceState(scope, {
    ...current,
    pinnedIds: toggleId(current.pinnedIds, sessionId),
  })
}

export function toggleWorkspaceSessionFavorite(scope: WorkspaceScope, sessionId: string) {
  const current = readWorkspacePreferenceState(scope)
  return writeWorkspacePreferenceState(scope, {
    ...current,
    favoriteIds: toggleId(current.favoriteIds, sessionId),
  })
}

export function isWorkspaceSessionPinned(state: WorkspaceSessionPreferenceState, sessionId: string) {
  return state.pinnedIds.includes(sessionId)
}

export function isWorkspaceSessionFavorite(state: WorkspaceSessionPreferenceState, sessionId: string) {
  return state.favoriteIds.includes(sessionId)
}

export function sortWorkspaceSessionsByPreference(
  items: WorkspaceSession[],
  state: WorkspaceSessionPreferenceState,
) {
  return items
    .map((item, index) => ({
      item,
      index,
      pinned: isWorkspaceSessionPinned(state, item.id),
      favorite: isWorkspaceSessionFavorite(state, item.id),
    }))
    .sort((left, right) => {
      if (left.pinned !== right.pinned) {
        return left.pinned ? -1 : 1
      }
      if (left.favorite !== right.favorite) {
        return left.favorite ? -1 : 1
      }
      if (left.item.status !== right.item.status) {
        return left.item.status === 'active' ? -1 : 1
      }
      return left.index - right.index
    })
    .map((entry) => entry.item)
}
