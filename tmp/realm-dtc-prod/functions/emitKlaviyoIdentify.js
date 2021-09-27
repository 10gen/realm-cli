exports = (email, customerProperties) => {
  const data = {
    properties: {
      $email: email,
      ...customerProperties,
    },
  };

  context.functions.execute('makeKlaviyoApiRequest', data, 'identify', email);
};
