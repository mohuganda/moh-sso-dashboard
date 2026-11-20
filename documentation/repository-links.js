import {statSync} from 'node:fs';
import {dirname, relative, resolve} from 'node:path';

export function repositoryLinks({repository, repositoryRoot}) {
  return (tree, file) => {
    function visit(node) {
      if (node.type === 'link' || node.type === 'definition') {
        const [pathname] = node.url.split(/[?#]/);
        if (pathname && !/^(?:[a-z][a-z\d+.-]*:|\/)/i.test(pathname)) {
          const target = resolve(dirname(file.path), decodeURIComponent(pathname));
          const repositoryPath = relative(repositoryRoot, target);
          const isGuide =
            /^(docs|frontend\/docs)\/.+\.mdx?$/.test(repositoryPath) ||
            repositoryPath === 'documentation/pages/index.md';

          if (!isGuide && repositoryPath && !repositoryPath.startsWith('../')) {
            const entry = statSync(target, {throwIfNoEntry: false});
            if (!entry) {
              throw new Error(`Missing repository link "${node.url}" in ${file.path}`);
            }
            const kind = entry.isDirectory() ? 'tree' : 'blob';
            const encodedPath = repositoryPath.split('/').map(encodeURIComponent).join('/');
            node.url = `${repository}/${kind}/main/${encodedPath}${node.url.slice(pathname.length)}`;
          }
        }
      }
      node.children?.forEach(visit);
    }

    visit(tree);
  };
}
