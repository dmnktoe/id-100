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
      version: window.APP_VERSION || 'dev',
      sessionSampleRate: 100,
      sessionReplaySampleRate: environment === 'production' ? 20 : 0, // No replay in dev
      trackBfcacheViews: true,
      defaultPrivacyLevel: 'mask-user-input',
    });

    // Set global context to distinguish frontend
    datadogRum.setGlobalContextProperty('layer', 'frontend');
    datadogRum.setGlobalContextProperty('platform', 'browser');

    console.log('Datadog RUM initialized successfully');
  } catch (error) {
    console.error('Failed to initialize Datadog RUM:', error);
  }
}
