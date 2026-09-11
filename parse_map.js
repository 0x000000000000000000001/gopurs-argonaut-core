const fs = require('fs');
const fn = JSON.parse(fs.readFileSync('output/Test.Main/corefn.json'));

function unwrap(node) {
    if (!node) return null;
    if (node.type === 'TypeApp') return unwrap(node.expression);
    if (node.type === 'Typed') return unwrap(node.expression);
    return node;
}

let count = 0;
function walk(node) {
  if (!node) return;
  let uNode = unwrap(node);
  if (!uNode) return;
  
  if (uNode.type === 'App') {
      let f1 = unwrap(uNode.abstraction);
      if (f1 && f1.type === 'App') {
          let f2 = unwrap(f1.abstraction);
          if (f2 && f2.type === 'App') {
              let f3 = unwrap(f2.abstraction);
              if (f3 && f3.type === 'Var' && f3.value.identifier === 'map') {
                  count++;
                  if (count === 1) { // The first map should be caseJsonNull
                      console.log("FOUND First map!");
                      console.log("Arg 2 (f):", JSON.stringify(uNode.abstraction.argument, null, 2));
                      console.log("Arg 3 (cases):", JSON.stringify(uNode.argument, null, 2));
                  }
              }
          }
      }
  }
  
  if (node.abstraction) walk(node.abstraction);
  if (node.argument) walk(node.argument);
  if (node.expression) walk(node.expression);
  if (node.body) walk(node.body);
  if (node.binds) node.binds.forEach(b => walk(b.expression));
  if (node.alternatives) node.alternatives.forEach(a => walk(a.expression));
}

fn.decls.forEach(d => walk(d.expression));
