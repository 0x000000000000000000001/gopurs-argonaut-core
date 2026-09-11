const fs = require('fs');
const json = JSON.parse(fs.readFileSync('foldXXX.json', 'utf8'));

function findMap(node) {
  if (Array.isArray(node)) {
    for (let x of node) {
      const res = findMap(x);
      if (res) return res;
    }
  } else if (node && typeof node === 'object') {
    if (node.type === 'App' && node.abstraction && node.abstraction.abstraction && node.abstraction.abstraction.value && node.abstraction.abstraction.value.identifier === 'map') {
      return node;
    }
    for (let key in node) {
      const res = findMap(node[key]);
      if (res) return res;
    }
  }
  return null;
}

const mapNode = findMap(json);
console.log(JSON.stringify(mapNode, null, 2));
