const fs = require('fs');
const json = JSON.parse(fs.readFileSync('foldXXX.json', 'utf8'));

function findArrayMap(node) {
  if (Array.isArray(node)) {
    for (let x of node) {
      const res = findArrayMap(x);
      if (res) return res;
    }
  } else if (node && typeof node === 'object') {
    if (node.type === 'App' && node.abstraction && node.abstraction.value && node.abstraction.value.identifier === 'arrayMap') {
      return node;
    }
    for (let key in node) {
      const res = findArrayMap(node[key]);
      if (res) return res;
    }
  }
  return null;
}

const mapNode = findArrayMap(json);
console.log(JSON.stringify(mapNode, null, 2));
