/* eslint-disable camelcase */
/* eslint-disable no-await-in-loop */
exports = async () => {
  const {
    REALM_PROJECT_ID,
    REALM_APP_ID,
  } = context.environment.values;
  if (!REALM_PROJECT_ID || !REALM_APP_ID) {
    console.log('Realm IDs unconfigured, log ingestion is disabled.');
    return;
  }

  const ADMIN_API_BASE_URL = context.values.get('ADMIN_API_BASE_URL');
  const LOGGER_PUBLIC_API_KEY = context.values.get('LOGGER_PUBLIC_API_KEY');
  const LOGGER_PRIVATE_API_KEY = context.values.get('LOGGER_PRIVATE_API_KEY');
  const logColl = context.services.get('mongodb-atlas')
    .db(context.values.get('DB_NAME'))
    .collection('realmlogs');
  const moment = require('moment');

  const { access_token } = await authenticate(
    ADMIN_API_BASE_URL,
    LOGGER_PUBLIC_API_KEY,
    LOGGER_PRIVATE_API_KEY,
  );
  const pager = new LogPager(ADMIN_API_BASE_URL, REALM_PROJECT_ID, REALM_APP_ID, access_token);

  const [lastLog] = await logColl.find({}).sort({ inserted_at: -1 }).limit(1).toArray();
  let hasNext = true;
  let prevPage = null;
  while (hasNext) {
    const page = await pager.getNextPage(prevPage);
    const logs = [];
    await Promise.all(page.logs.map(async (log) => {
      try {
        log.end_date = page.nextEndDate;
        log.skip = page.nextSkip;
        log.inserted_at = new Date();
        await logColl.insertOne(log);
        logs.push(log);
      } catch (e) {
        if (!e.message.includes('E11000')) {
          console.error(`Error creating log ${JSON.stringify(log._id)}`, e);
        }
      }
    }));
    if (logs.length > 0) {
      await context.functions.execute('forwardLogs', logs);
    } else {
      return;
    }
    if (moment(page.nextEndDate).isBefore(lastLog.end_date)) {
      return;
    }
    hasNext = page.nextEndDate;
    prevPage = page;
  }
};

class LogPager {
  constructor(adminBaseUrl, projectId, appId, access_token, queryParams = {}) {
    this.logsEndpoint = `${adminBaseUrl}/groups/${projectId}/apps/${appId}/logs`;
    this.queryParams = queryParams;
    this.authHeaders = { Authorization: [`Bearer ${access_token}`] };
  }

  async getNextPage(prevPage) {
    const { nextEndDate, nextSkip } = prevPage || {};
    if (prevPage && !nextEndDate) {
      throw new Error('Paginated API does not have any more pages.');
    }
    const request = {
      headers: this.authHeaders,
      url: this.logsEndpoint + formatQueryString({
        ...this.queryParams,
        end_date: nextEndDate,
        skip: nextSkip,
      }),
    };
    const result = await context.http.get(request);
    const nextPage = EJSON.parse(result.body.text());
    return nextPage;
  }
}

async function authenticate(adminBaseUrl, publicApiKey, privateApiKey) {
  const result = await context.http.post({
    url: `${adminBaseUrl}/auth/providers/mongodb-cloud/login`,
    headers: {
      'Content-Type': ['application/json'],
      Accept: ['application/json'],
    },
    body: {
      username: publicApiKey,
      apiKey: privateApiKey,
    },
    encodeBodyAsJSON: true,
  });
  if (result.statusCode !== 200) {
    throw new Error(`Error authenticating log forwarder: ${JSON.stringify(result)}`);
  }
  return EJSON.parse(result.body.text());
}

function formatQueryString(queryParams) {
  const params = Object.entries(queryParams).filter(([, b]) => b);
  return params.length > 0
    ? `?${params.map(([a, b]) => `${a}=${b}`).join('&')}`
    : '';
}
