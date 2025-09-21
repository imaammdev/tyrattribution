// Shared utilities for k6 tests

export function generateUUID() {
  return 'xxxxxxxx-xxxx-4xxx-yxxx-xxxxxxxxxxxx'.replace(/[xy]/g, function(c) {
    const r = Math.random() * 16 | 0;
    const v = c === 'x' ? r : (r & 0x3 | 0x8);
    return v.toString(16);
  });
}

export const CAMPAIGN_IDS = [
  '550e8400-e29b-41d4-a716-446655440000',
  '6ba7b810-9dad-11d1-80b4-00c04fd430c8',
  '6ba7b811-9dad-11d1-80b4-00c04fd430c8',
];

export const SOURCES = ['google', 'facebook', 'twitter', 'instagram'];
export const CONVERSION_TYPES = ['purchase', 'signup', 'download', 'subscription'];

export function getRandomItem(array) {
  return array[Math.floor(Math.random() * array.length)];
}

export const BASE_URL = __ENV.BASE_URL || 'http://localhost:8080';
export const CLICK_ENDPOINT = `${BASE_URL}/api/clicks`;
export const CONVERSION_ENDPOINT = `${BASE_URL}/api/conversions`;