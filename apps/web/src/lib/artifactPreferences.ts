import type { WorkspaceScope } from './types'

export interface ArtifactPreferenceState {
  favoriteIds: string[]
  recentIds: string[]
}

const artifactPreferenceStoragePrefix = 'openclaw.artifact-preferences'
const artifactRecentLimit = 12

function artifactPreferenceStorageKey(scope: WorkspaceScope) {
  return `${artifactPreferenceStoragePrefix}.${scope}`
}

function normalizeIdList(value: unknown) {
  if (!Array.isArray(value)) {
    return []
  }
  return value
    .map((item) => (typeof item === 'string' ? item.trim() : ''))
    .filter((item, index, items) => item && items.indexOf(item) === index)
}

function readArtifactPreferenceState(scope: WorkspaceScope): ArtifactPreferenceState {
  if (typeof window === 'undefined') {
    return { favoriteIds: [], recentIds: [] }
  }

  try {
    const raw = window.localStorage.getItem(artifactPreferenceStorageKey(scope))
    if (!raw) {
      return { favoriteIds: [], recentIds: [] }
    }

    const parsed = JSON.parse(raw) as { favoriteIds?: unknown; recentIds?: unknown }
    return {
      favoriteIds: normalizeIdList(parsed.favoriteIds),
      recentIds: normalizeIdList(parsed.recentIds),
    }
  } catch {
    return { favoriteIds: [], recentIds: [] }
  }
}

function writeArtifactPreferenceState(scope: WorkspaceScope, state: ArtifactPreferenceState) {
  if (typeof window === 'undefined') {
    return state
  }

  const normalized = {
    favoriteIds: normalizeIdList(state.favoriteIds),
    recentIds: normalizeIdList(state.recentIds).slice(0, artifactRecentLimit),
  }

  try {
    window.localStorage.setItem(artifactPreferenceStorageKey(scope), JSON.stringify(normalized))
  } catch {
    // ignore storage failures
  }

  return normalized
}

function toggleId(ids: string[], id: string) {
  return ids.includes(id) ? ids.filter((item) => item !== id) : [id, ...ids]
}

export function loadArtifactPreferences(scope: WorkspaceScope) {
  return readArtifactPreferenceState(scope)
}

export function toggleArtifactFavorite(scope: WorkspaceScope, artifactId: string) {
  const current = readArtifactPreferenceState(scope)
  return writeArtifactPreferenceState(scope, {
    ...current,
    favoriteIds: toggleId(current.favoriteIds, artifactId),
  })
}

export function markArtifactRecentlyViewed(scope: WorkspaceScope, artifactId: string) {
  const current = readArtifactPreferenceState(scope)
  const recentIds = [artifactId, ...current.recentIds.filter((item) => item !== artifactId)].slice(0, artifactRecentLimit)
  return writeArtifactPreferenceState(scope, {
    ...current,
    recentIds,
  })
}

export function isArtifactFavorite(state: ArtifactPreferenceState, artifactId: string) {
  return state.favoriteIds.includes(artifactId)
}

export function selectArtifactsByIdOrder<T extends { id: string }>(items: T[], ids: string[], limit = ids.length) {
  const itemMap = new Map(items.map((item) => [item.id, item]))
  const selected: T[] = []
  for (const id of ids) {
    const match = itemMap.get(id)
    if (match) {
      selected.push(match)
    }
    if (selected.length >= limit) {
      break
    }
  }
  return selected
}

export function sortArtifactsByPreference<T extends { id: string }>(items: T[], state: ArtifactPreferenceState) {
  return items
    .map((item, index) => ({
      item,
      index,
      favorite: isArtifactFavorite(state, item.id),
    }))
    .sort((left, right) => {
      if (left.favorite !== right.favorite) {
        return left.favorite ? -1 : 1
      }
      return left.index - right.index
    })
    .map((entry) => entry.item)
}
