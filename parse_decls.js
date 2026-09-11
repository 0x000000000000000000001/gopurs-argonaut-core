const fs = require('fs');
const json = JSON.parse(fs.readFileSync('output/Test.Main/corefn.json', 'utf8'));

for (let d of json.decls) {
  if (d.type === 'NonRec') {
    console.log("NonRec:", d.identifier);
  } else if (d.type === 'Rec') {
    for (let b of d.binds) {
      console.log("Rec:", b.identifier);
    }
  }
}
