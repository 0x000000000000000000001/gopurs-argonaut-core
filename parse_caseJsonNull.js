const fs = require('fs');
const json = JSON.parse(fs.readFileSync('output/Test.Main/corefn.json', 'utf8'));

for (let d of json.decls) {
  if (d.type === 'NonRec' && d.expression && d.expression.type === 'App') {
    // Check if this binds caseJsonNull
    console.log("Binding:", d.identifier);
  }
}
