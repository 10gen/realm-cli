/*
  This function is run when a GraphQL Query is made requesting your
  custom field name. The return value of this function is used to
  populate the resolver generated from your Payload Type.

  This function expects the following input object:

  {
    "type": "object",
    "title": "PopulateInput",
    "required": [
      "sourceField"
    ],
    "properties": {
      "limit": {
        "bsonType": "int"
      },
      "skip": {
        "bsonType": "int"
      },
      "sort": {
        "bsonType": "int"
      },
      "sourceField": {
        "bsonType": "string"
      },
      "targetField": {
        "bsonType": "string"
      }
    }
  }
*/

exports = async (input, doc) => {
  input.collection = 'messages';
  return context.functions.execute('populate', input, doc);
};
