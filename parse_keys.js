const fs = require('fs');
const json = JSON.parse(fs.readFileSync('output/Test.Main/corefn.json', 'utf8'));
console.log(Object.keys(json));
