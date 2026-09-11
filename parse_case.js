const fs = require('fs');
const json = JSON.parse(fs.readFileSync('output/Test.Main/corefn.json', 'utf8'));

function findCase(node) {
  if (Array.isArray(node)) {
    for (let x of node) findCase(x);
  } else if (node && typeof node === 'object') {
    if (node.type === 'Var' && node.value && node.value.identifier === 'caseJsonNull') {
      console.log('FOUND caseJsonNull');
    }
    for (let key in node) findCase(node[key]);
  }
}

findCase(json.decls);
