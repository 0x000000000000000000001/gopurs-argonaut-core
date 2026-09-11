const fs = require('fs');
const json = JSON.parse(fs.readFileSync('foldXXX.json', 'utf8'));

function findLets(node) {
  if (Array.isArray(node)) {
    for (let x of node) findLets(x);
  } else if (node && typeof node === 'object') {
    if (node.type === 'Let') {
      for (let b of node.binds) {
        console.log("Let bind:", b.identifier);
      }
    }
    for (let key in node) findLets(node[key]);
  }
}

findLets(json);
