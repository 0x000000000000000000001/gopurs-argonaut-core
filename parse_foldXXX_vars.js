const fs = require('fs');
const json = JSON.parse(fs.readFileSync('foldXXX.json', 'utf8'));

function findVars(node, vars) {
  if (Array.isArray(node)) {
    for (let x of node) findVars(x, vars);
  } else if (node && typeof node === 'object') {
    if (node.type === 'Var' && node.value && node.value.identifier) {
      vars.add(node.value.identifier);
    }
    for (let key in node) {
      findVars(node[key], vars);
    }
  }
}

const vars = new Set();
findVars(json, vars);
console.log(Array.from(vars));
