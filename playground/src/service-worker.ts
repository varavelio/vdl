// Disables access to DOM typings like `HTMLElement` which are not available
// inside a service worker and instantiates the correct globals
/// <reference no-default-lib="true"/>
/// <reference lib="esnext" />
/// <reference lib="webworker" />
//
// Ensures that the `$service-worker` import has proper type definitions
/// <reference types="@sveltejs/kit" />
//
// Only necessary if you have an import from `$env/static/public`
/// <reference types="../.svelte-kit/ambient.d.ts" />
/**
 * SvelteKit Service Worker
 *
 * Docs: https://svelte.dev/docs/kit/service-workers
 *
 * Caching criteria:
 * - Caches are scoped to the service worker's path.
 * - Caches same-origin assets ONLY if they are inside an `/_app/` subdirectory
 *   relative to the service worker's scope.
 * - Caches cross-origin assets (CDNs, etc.).
 * - Caches only GET requests.
 *
 * Caching strategy:
 * A "Cache First, falling back to Network" strategy is used for all
 * assets that meet the caching criteria.
 */
import { version } from "$service-worker";

const swSelf = globalThis.self as unknown as ServiceWorkerGlobalScope;

/**
 * Generates the cache key prefix for the given scope.
 * @param scope The service worker's registration scope.
 * @returns  The cache key prefix for the given scope.
 */
const getCacheKeyPrefix = (scope: string): string => {
  return `sw-cache-${scope}`;
};

/**
 * Generates the dynamic cache key for the current scope and version.
 * @param {string} scope - The service worker's registration scope.
 * @returns {string} The fully-formed cache key.
 */
const getCacheKey = (scope: string): string => {
  return `${getCacheKeyPrefix(scope)}-${version}`;
};

swSelf.addEventListener("install", (event: ExtendableEvent) => {
  event.waitUntil(swSelf.skipWaiting());
});

swSelf.addEventListener("activate", (event: ExtendableEvent) => {
  const scope = swSelf.registration.scope;
  const currentCacheKey = getCacheKey(scope);

  async function cleanOldCaches() {
    const cacheNames = await caches.keys();

    for (const cacheName of cacheNames) {
      if (cacheName === currentCacheKey) continue;
      if (cacheName.startsWith(getCacheKeyPrefix(scope))) {
        await caches.delete(cacheName);
      }
    }

    await swSelf.clients.claim();
  }

  event.waitUntil(cleanOldCaches());
});

swSelf.addEventListener("fetch", (event: FetchEvent) => {
  const scope = swSelf.registration.scope;
  const request = event.request;

  const scopeURL = new URL(scope);
  const requestURL = new URL(request.url);

  const scopePath = scopeURL.pathname;
  const requestPath = requestURL.pathname;

  const isSameOrigin = requestURL.origin === scopeURL.origin;
  const isInAppPath = requestPath.startsWith(`${scopePath}_app/`);

  if (request.method !== "GET") return;
  if (isSameOrigin && !isInAppPath) return;

  async function respondCacheFirst(): Promise<Response> {
    const currentCacheKey = getCacheKey(scope);
    const cache = await caches.open(currentCacheKey);
    const cachedResponse = await cache.match(request);

    // Cache first
    if (cachedResponse) return cachedResponse;

    // Network fallback (caches if successful)
    try {
      const networkResponse = await fetch(request);
      if (networkResponse && networkResponse.status === 200) {
        await cache.put(request, networkResponse.clone());
      }
      return networkResponse;
    } catch (error) {
      console.error(`[SW] Network fetch failed for: ${request.url}`, error);
      return new Response(`Network error for ${request.url}`, {
        status: 408,
        headers: { "Content-Type": "text/plain" },
      });
    }
  }

  event.respondWith(respondCacheFirst());
});
