/** 动态路由是否已为当前会话加载（切换用户须重置） */
let isRoutesLoaded = false
/** 并发导航共享同一次动态路由加载，避免重复 reset/addRoute */
let routesLoadingPromise: Promise<void> | null = null

export function isRoutesLoadedState() {
  return isRoutesLoaded
}

export function markRoutesLoaded() {
  isRoutesLoaded = true
}

export function resetRoutesLoadedFlag() {
  isRoutesLoaded = false
  routesLoadingPromise = null
}

/** 确保动态路由只加载一次；并发 beforeEach 会 await 同一 Promise */
export async function ensureRoutesLoaded(load: () => Promise<unknown>) {
  if (isRoutesLoaded)
    return
  if (!routesLoadingPromise) {
    routesLoadingPromise = load()
      .then(() => {
        markRoutesLoaded()
      })
      .finally(() => {
        routesLoadingPromise = null
      })
  }
  await routesLoadingPromise
}
