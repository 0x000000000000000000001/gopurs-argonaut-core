const fs = require('fs');
const json = JSON.parse(fs.readFileSync('output/Test.Main/corefn.json', 'utf8'));

for (let d of json.decls) {
  if (d.identifier === 'foldXXX') {
    console.log(JSON.stringify(d, null, 2));
  }
}
