exports = async (event) => {
  try {
    const EVENT_ROUTING_KEY = context.values.get('VERB_EVENT_ROUTING_KEY');
    const message = {
      timestamp: new Date(),
      ...event,
    };
    if (!message.source) {
      message.source = 'REALM';
    }
    if (EVENT_ROUTING_KEY) {
      await context.functions.execute('publishMessage', message, EVENT_ROUTING_KEY);
    } else {
      console.log(`Event routing unconfigured for message ${JSON.stringify(message)}`);
    }
  } catch (e) {
    console.error(`Event emission failed for ${JSON.stringify(event)}`, e);
  }
};
