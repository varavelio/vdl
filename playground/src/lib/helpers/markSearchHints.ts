import type { SearchResult } from "minisearch";

/**
 * Highlights matching terms in a text string by wrapping them in HTML mark tags
 *
 * @param searchResult - The search result containing the terms to highlight
 * @param textToMark - The text string where terms should be highlighted
 * @returns The text string with matching terms wrapped in <mark> tags
 */
export function markSearchHints(searchResult: string[], textToMark: string) {
  // Escape special regex characters in each search term to avoid regexp errors
  const escapeRegExp = (str: string) => str.replace(/[.*+?^${}()|[\]\\]/g, "\\$&");

  // Filter out empty strings to avoid empty alternatives in the regexp
  const sanitizedTerms = searchResult.filter((term) => term && term.length > 0).map(escapeRegExp);

  if (sanitizedTerms.length === 0) {
    return textToMark;
  }

  const regexp = new RegExp(`(${sanitizedTerms.join("|")})`, "gi");
  return textToMark.replace(regexp, "<mark>$1</mark>");
}

/**
 * Truncates text around marked terms, showing context before the match
 *
 * @param searchResult - The search result containing the terms to highlight
 * @param textToMark - The text string to truncate and highlight
 * @returns The truncated text with matching terms highlighted. If the match is within
 *          the first 3 words, returns the full text. Otherwise, shows 3 words before
 *          the match and everything after, with an ellipsis prefix.
 */
export function truncateWithMark(searchResult: string[], textToMark: string) {
  // First mark the text
  const markedText = markSearchHints(searchResult, textToMark);

  // Check if there's a mark in the text
  if (!markedText.includes("<mark>")) {
    return markedText;
  }

  // Get the position of the first mark
  const firstMarkIndex = markedText.indexOf("<mark>");
  const textBeforeMark = markedText.substring(0, firstMarkIndex);

  // Count words before the mark
  const wordsBeforeMark = textBeforeMark.trim().split(/\s+/).length;

  // If mark is within first 3 words, return as is
  if (wordsBeforeMark <= 3) {
    return markedText;
  }

  // If mark is after 3 words, truncate and add ellipsis
  const words = markedText.split(/\s+/);
  const markIndex = words.findIndex((word) => word.includes("<mark>"));

  // Get the 3 words before the mark and the marked text
  const startIndex = Math.max(0, markIndex - 3);
  const truncatedWords = words.slice(startIndex);

  return `... ${truncatedWords.join(" ")}`;
}

/**
 * Same as markSearchHints but for minisearch
 *
 * @param searchResult - The search result containing the terms to highlight
 * @param textToMark - The text string where terms should be highlighted
 * @returns The text string with matching terms wrapped in <mark> tags
 */
export function markSearchHintsMinisearch(searchResult: SearchResult, textToMark: string) {
  return markSearchHints(searchResult.terms, textToMark);
}

/**
 * Same as truncateWithMark but for minisearch
 *
 * @param searchResult - The search result containing the terms to highlight
 * @param textToMark - The text string to truncate and highlight
 * @returns The truncated text with matching terms highlighted. If the match is within
 *          the first 3 words, returns the full text. Otherwise, shows 3 words before
 *          the match and everything after, with an ellipsis prefix.
 */
export function truncateWithMarkMinisearch(searchResult: SearchResult, textToMark: string) {
  return truncateWithMark(searchResult.terms, textToMark);
}
