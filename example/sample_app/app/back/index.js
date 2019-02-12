const express = require('express');
const app = express();
const bodyParser = require('body-parser');
const apiMetrics = require('prometheus-api-metrics');

app.use(apiMetrics())

app.use(bodyParser.json()); // for parsing application/json

app.post('/', (req, res) => {

  if (!req.body) return res.sendStatus(400);

  data = JSON.stringify(req.body);

  console.log("received from front end " + data + "; echoing back");

  res.send(data);
})

app.listen(process.env.APP_PORT, () => {
  console.log('listening on ' + process.env.APP_PORT)
});
