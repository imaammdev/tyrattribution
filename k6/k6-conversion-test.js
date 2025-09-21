import http from 'k6/http';
import { check, sleep } from 'k6';
import { Rate } from 'k6/metrics';
import { generateUUID, CAMPAIGN_IDS, SOURCES, CONVERSION_TYPES, getRandomItem, CONVERSION_ENDPOINT } from './utils.js';

// Custom metrics
export let errorRate = new Rate('errors');
export let conversionRate = new Rate('successful_conversions');

// Test configuration for conversion scenarios
export let options = {
  executor: 'constant-vus',
  vus: 15,
  duration: '5m',
  thresholds: {
    http_req_duration: ['p(95)<1000'],
    http_req_failed: ['rate<0.05'],
    errors: ['rate<0.05'],
  },
};

// Conversion test scenario
export default function() {
  const userId = generateUUID();
  const campaignId = getRandomItem(CAMPAIGN_IDS);
  const source = getRandomItem(SOURCES);

  // Create conversion event
  const conversionTime = new Date();
  const conversionPayload = {
    conversion_id: generateUUID(),
    user_id: userId,
    campaign_id: campaignId,
    conversion_date: conversionTime.toISOString(),
    value: Math.random() * 200 + 50, // $50-$250
    type: getRandomItem(CONVERSION_TYPES),
    source: source
  };

  const params = {
    headers: { 'Content-Type': 'application/json' },
  };

  const conversionResponse = http.post(CONVERSION_ENDPOINT, JSON.stringify(conversionPayload), params);

  const conversionSuccess = check(conversionResponse, {
    'conversion created successfully': (r) => r.status === 201,
    'conversion response has conversion_id': (r) => {
      try {
        return JSON.parse(r.body).conversion_id !== undefined;
      } catch (e) {
        return false;
      }
    },
  });

  if (conversionSuccess) {
    conversionRate.add(1);
    errorRate.add(0);
  } else {
    conversionRate.add(0);
    errorRate.add(1);

    // Log error details in real-time
    const timestamp = new Date().toISOString();
    console.error(`[${timestamp}] CONVERSION ERROR - Status: ${conversionResponse.status}, Body: ${conversionResponse.body}, Campaign: ${campaignId}, User: ${userId}, Value: ${conversionPayload.value}`);
  }

  // Wait random time (1-3 seconds) between requests
  sleep(Math.random() * 2 + 1);
}