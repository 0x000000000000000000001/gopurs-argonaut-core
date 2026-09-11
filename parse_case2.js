const fs = require('fs');
const json = JSON.parse(fs.readFileSync('output/Test.Main/corefn.json', 'utf8'));

function findCase(node, parent) {
  if (Array.isArray(node)) {
    for (let x of node) findCase(x, parent);
  } else if (node && typeof node === 'object') {
    if (node.type === 'Var' && node.value && node.value.identifier === 'caseJsonNull') {
      console.log(JSON.stringify(parent, null, 2));
    }
    for (let key in node) findCase(node[key], node);
  }
}

findCase(json.decls, null);
