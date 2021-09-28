/*
  MongoDB Realm function for creating a stripe charge for a Verb order

  input: {
    customer - the customer to charge
    order - the order to be charged
  }

  returns: the Stripe charge response
*/

exports = async (input) => {
  try {
    const STRIPE_SECRET_KEY = context.values.get('STRIPE_SECRET_KEY');
    const { http } = context;
    const querystring = require('querystring');

    const { order, customer } = input;
    const customerId = customer.stripeID || customer.customerId;
    const formData = {
      amount: order.paid.total,
      customer: customerId,
      description: order.invoiceNumber,
      'metadata[invoiceNumber]': order.invoiceNumber,
      currency: 'usd',
    };

    const headers = {
      Authorization: [`Bearer ${STRIPE_SECRET_KEY}`],
      'Content-Type': ['application/x-www-form-urlencoded'],
    };
    const body = querystring.stringify(formData);
    const response = await http.post({
      url: 'https://api.stripe.com/v1/charges',
      body,
      headers,
    });
    const responseBody = EJSON.parse(response.body.text());
    if (response.statusCode !== 200) {
      throw new Error(`stripe charge request ${JSON.stringify(formData)} failed with error: ${JSON.stringify(responseBody)}`);
    }
    return responseBody;
  } catch (e) {
    e.code = 'FAILED_CHARGE';
    throw e;
  }
};
