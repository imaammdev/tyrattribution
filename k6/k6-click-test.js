import http from 'k6/http';
import { check, sleep } from 'k6';
import { Rate } from 'k6/metrics';
import { generateUUID, CAMPAIGN_IDS, SOURCES, getRandomItem, CLICK_ENDPOINT } from './utils.js';

// Custom metrics
export let errorRate = new Rate('errors');
export let clickRate = new Rate('successful_clicks');

// Test configuration for click scenarios
export let options = {
  executor: 'constant-vus',
  vus: 10000,
  duration: '1m',
  thresholds: {
    http_req_duration: ['p(95)<1000'],
    http_req_failed: ['rate<0.05'],
    errors: ['rate<0.05'],
  },
};

// Click test scenario
export default function() {
  const userId = generateUUID();
  const campaignId = getRandomItem(CAMPAIGN_IDS);
  const source = getRandomItem(SOURCES);

  // Create click event
  const clickTime = new Date();
  const clickPayload = {
    campaign_id: campaignId,
    user_id: userId,
    click_date: clickTime.toISOString(),
    source: source
  };

  const params = {
    headers: { 'Content-Type': 'application/json' },
  };

  const clickResponse = http.post(CLICK_ENDPOINT, JSON.stringify(clickPayload), params);

  const clickSuccess = check(clickResponse, {
    'click created successfully': (r) => r.status === 201,
    'click response has click_id': (r) => {
      try {
        return JSON.parse(r.body).click_id !== undefined;
      } catch (e) {
        return false;
      }
    },
  });

  if (clickSuccess) {
    clickRate.add(1);
    errorRate.add(0);
  } else {
    clickRate.add(0);
    errorRate.add(1);

    // Log error details in real-time
    const timestamp = new Date().toISOString();
    console.error(`[${timestamp}] CLICK ERROR - Status: ${clickResponse.status}, Body: ${clickResponse.body}, Campaign: ${campaignId}, User: ${userId}`);
  }

  // Wait random time (1-3 seconds) between requests
  sleep(Math.random() * 2 + 1);
}