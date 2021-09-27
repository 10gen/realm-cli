/*
  This function is run when a GraphQL Query is made requesting your
  custom field name. The return value of this function is used to
  populate the resolver generated from your Payload Type.

  This function expects the following input object:

  {
    "type": "object",
    "title": "populateInput",
    "required": [
      "collection",
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
      "collection": {
        "bsonType": "string"
      },
      "sourceField": {
        "bsonType": "string"
      },
      "targetField": {
        "bsonType": "string"
      }
    }
  }

  And the following payload object:

  {
    "type": "object",
    "title": "populateResult",
    "properties": {
      "resultType": {
        "bsonType": "string"
      },
      "populated": {
        "bsonType": "array",
        "items": {
          "bsonType": "object"
        }
      }
    }
  }
*/

exports = async (input) => {
  const cluster = context.services.get('mongodb-atlas');
  const targetColl = cluster.db('verbenergy').collection('tickets');
  return targetColl.count(input);
};
