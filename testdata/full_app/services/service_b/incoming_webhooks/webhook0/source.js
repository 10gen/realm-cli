var exports = function(payload) {
  /*
    See https://developer.github.com/v3/activity/events/types/#webhook-payload-example for
    examples of payloads.

    Try running this in the console below:
    exports({pull_request: {}, action: 'opened'})
  */
  return payload.pull_request && payload.action === 'opened';
};