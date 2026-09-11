const fs = require('fs');

function walk(node, file) {
  if (!node) return;
  if (node.type === 'Var' && node.value && node.value.identifier === 'arrayMap') {
      console.log(`FOUND arrayMap Var in ${file}!`);
  }
  
  if (node.type === 'App') {
      walk(node.abstraction, file);
      walk(node.argument, file);
  } else if (node.type === 'Typed') {
      walk(node.expression, file);
  } else {
      if (node.expression) walk(node.expression, file);
      if (node.body) walk(node.body, file);
      if (node.binds) node.binds.forEach(b => walk(b.expression, file));
      if (node.alternatives) node.alternatives.forEach(a => walk(a.expression, file));
  }
}

const file = process.argv[2];
const fn = JSON.parse(fs.readFileSync(file));
fn.decls.forEach(d => walk(d.expression, file));
