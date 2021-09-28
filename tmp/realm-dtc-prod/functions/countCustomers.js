/*
  This function is run when a GraphQL Query is made requesting your
  custom field name. The return value of this function is used to
  populate the resolver generated from your Payload Type.

  This function expects the following input object:

  {
    "type": "object",
    "title": "CustomerQueryInput",
    "properties": {
      "query": {
        "type": "string"
      }
    },
    "required": ["query"]
  }

  And returns the following payload object:

  {
    "type": "object",
    "title": "countCustomersResult",
    "properties": {
      "count": {
        "type": "number"
      }
    }
  }
*/

exports = async (input) => {
  const cluster = context.services.get('mongodb-atlas');
  const targetColl = cluster.db('verbenergy').collection('newcustomers');
  return targetColl.count(input);
};
