/**
 * Datadog RUM (Real User Monitoring) integration for frontend
 */

import { datadogRum } from '@datadog/browser-rum';

/**
 * Initialize Datadog RUM tracking
 */
export function initDatadog(): void {
  const applicationId = window.DATADOG_APP_ID;
  const clientToken = window.DATADOG_CLIENT_TOKEN;
  const environment = window.ENVIRONMENT || 'development';
  
  if (!applicationId || !clientToken) {
    console.log('Datadog not configured, skipping Datadog RUM initialization');
    return;
  }

  try {
    datadogRum.init({
      applicationId,
      clientToken,
      site: 'datadoghq.eu',
      service: 'id-100',
      env: environment,
      // Version will be added later if needed
      // version: '1.0.0',
      sessionSampleRate: 100,
      sessionReplaySampleRate: environment === 'production' ? 20 : 0, // No replay in dev
      trackBfcacheViews: true,
      defaultPrivacyLevel: 'mask-user-input',
    });

    console.log('Datadog RUM initialized successfully');
  } catch (error) {
    console.error('Failed to initialize Datadog RUM:', error);
  }
}
