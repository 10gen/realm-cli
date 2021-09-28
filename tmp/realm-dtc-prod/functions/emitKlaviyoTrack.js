exports = async (email, event, properties) => {
  const data = {
    event,
    customer_properties: {
      $email: email,
    },
    properties,
  };
  context.functions.execute('makeKlaviyoApiRequest', data, 'track', email);
};
