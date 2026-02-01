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
 * Caching strategy:
 * A "Cache First, falling back to Network" strategy is used for all
 * network requests with the exception of all requests that terminate
 * with any of the files listed in NETWORK_FIRST_FILES, those use
 * a "Network First, falling back to Cache" strategy.
 *
 * The "version" import changes in every build, use it to invalidate cache on updates.
 */
import { build, files, prerendered, version } from "$service-worker";

const swSelf = globalThis.self as unknown as ServiceWorkerGlobalScope;
const KNOWN_ASSETS = [...build, ...files, ...prerendered];
const NETWORK_FIRST_FILES = [
  "/",
  "index.html",
  "version.json",
  "config.json",
  "schema.vdl",
  "openapi.yaml",
  "openapi.yml",
];

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

swSelf.addEventListener("install", (event) => {
  const scope = swSelf.registration.scope;
  const currentCacheKey = getCacheKey(scope);
  const cleanKnownAssets = KNOWN_ASSETS.map((asset) => {
    // If is URL return as is
    if (asset.startsWith("http")) return asset;
    // Make the path relative by removing the first /
    const relativePath = asset.startsWith("/") ? asset.slice(1) : asset;
    // Create an absolute path using the current scope
    return new URL(relativePath, scope).href;
  });

  async function addFilesToCache() {
    const cache = await caches.open(currentCacheKey);
    await cache.addAll(cleanKnownAssets);
  }

  event.waitUntil(addFilesToCache());
  swSelf.skipWaiting();
});

swSelf.addEventListener("activate", (event: ExtendableEvent) => {
  const scope = swSelf.registration.scope;
  const cachePrefix = getCacheKeyPrefix(scope);
  const currentCacheKey = getCacheKey(scope);

  async function cleanOldCaches() {
    const cacheNames = await caches.keys();

    for (const cacheName of cacheNames) {
      if (cacheName === currentCacheKey) continue;
      if (cacheName.startsWith(cachePrefix)) {
        await caches.delete(cacheName);
      }
    }

    await swSelf.clients.claim();
  }

  event.waitUntil(cleanOldCaches());
});

swSelf.addEventListener("fetch", (event: FetchEvent) => {
  const request = event.request;
  if (request.method !== "GET") return;

  async function respondNetworkFirst(): Promise<Response> {
    const currentCacheKey = getCacheKey(swSelf.registration.scope);
    const cache = await caches.open(currentCacheKey);

    try {
      // Try network first
      const networkResponse = await fetch(request);
      if (networkResponse.status === 200) {
        await cache.put(request, networkResponse.clone());
      }
      return networkResponse;
    } catch (error) {
      // Fallback to cache (Offline)
      const cachedResponse = await cache.match(request);
      if (cachedResponse) return cachedResponse;
      throw error;
    }
  }

  async function respondCacheFirst(): Promise<Response> {
    const currentCacheKey = getCacheKey(swSelf.registration.scope);
    const cache = await caches.open(currentCacheKey);
    const cachedResponse = await cache.match(request);

    // Try cache first
    if (cachedResponse) return cachedResponse;

    // Fallback to network (caches if successful)
    try {
      const networkResponse = await fetch(request);
      if (
        networkResponse &&
        (networkResponse.status === 200 || networkResponse.status === 0)
      ) {
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

  const url = new URL(request.url);
  const isNavigation = request.mode === "navigate";
  const isCriticalFile = NETWORK_FIRST_FILES.some((file) =>
    url.pathname.endsWith(file),
  );

  if (isNavigation || isCriticalFile) {
    event.respondWith(respondNetworkFirst());
    return;
  }

  event.respondWith(respondCacheFirst());
});
