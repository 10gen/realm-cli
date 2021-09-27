exports = async (klaviyoApiPayload, requestType = 'track', email = '') => {
  try {
    const KLAVIYO_KEY = context.values.get('KLAVIYO_KEY');
    const KLAVIYO_API_BASE_URL = context.values.get('KLAVIYO_API_BASE_URL');
    const { KLAVIYO_ENABLED } = context.environment.values;
    const KLAVIYO_BYPASS_EMAIL_SUFFIX = 'verbenergy.co';

    const { http } = context;
    klaviyoApiPayload.token = KLAVIYO_KEY;
    const encodedData = Buffer.from(JSON.stringify(klaviyoApiPayload)).toString('base64');
    if (KLAVIYO_ENABLED || email.endsWith(KLAVIYO_BYPASS_EMAIL_SUFFIX)) {
      const response = await http.get({ url: `${KLAVIYO_API_BASE_URL}${requestType}?data=${encodedData}` });
      if (response.statusCode !== 200) {
        throw new Error(`Klaviyo request failed for encoded data ${encodedData}`);
      }
    } else {
      console.log(`Klaviyo disabled for ${email}. Bypassing request with encoded data ${encodedData}`);
    }
  } catch (e) {
    console.error('Klaviyo request failed', e);
  }
};
