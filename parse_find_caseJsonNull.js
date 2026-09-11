const fs = require('fs');
const json = JSON.parse(fs.readFileSync('output/Test.Main/corefn.json', 'utf8'));

function findCase(node) {
  if (Array.isArray(node)) {
    for (let x of node) {
      const res = findCase(x);
      if (res) return res;
    }
  } else if (node && typeof node === 'object') {
    if (node.type === 'Var' && node.value && node.value.identifier === 'caseJsonNull') {
      return true;
    }
    for (let key in node) {
      const res = findCase(node[key]);
      if (res) return true;
    }
  }
  return false;
}

for (let d of json.decls) {
  if (findCase(d)) {
    console.log("CONTAINS caseJsonNull:", d.identifier);
  }
}
