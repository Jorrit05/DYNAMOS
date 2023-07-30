import json

data1 = {
  "Aadnstvb": {
    "listValue": {
      "values": [
        {
          "stringValue": "V"
        }
      ]
    }
  }
}

data2 = {
  "Aadnstvb": ["V"]
}

json_data1 = json.dumps(data1)
json_data2 = json.dumps(data2)

print(len(json_data1.encode('utf-8')))  # Prints the byte size of data1
print(len(json_data2.encode('utf-8')))