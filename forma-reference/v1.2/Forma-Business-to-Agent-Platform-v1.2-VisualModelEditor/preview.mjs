import http from 'node:http';
import {readFile,stat} from 'node:fs/promises';
import path from 'node:path';
import {fileURLToPath} from 'node:url';
const root=path.resolve(fileURLToPath(new URL('./preview-dist/',import.meta.url)));
const mime={'.html':'text/html; charset=utf-8','.js':'text/javascript; charset=utf-8','.css':'text/css; charset=utf-8','.svg':'image/svg+xml','.png':'image/png','.json':'application/json'};
http.createServer(async(req,res)=>{try{const pathname=decodeURIComponent(new URL(req.url,'http://localhost').pathname);let file=path.resolve(root,'.'+pathname);if(file!==root&&!file.startsWith(root+path.sep)){res.writeHead(403);res.end();return}try{if(!(await stat(file)).isFile())file=path.join(root,'index.html')}catch{if(path.extname(file)){res.writeHead(404);res.end('Not found');return}file=path.join(root,'index.html')}res.writeHead(200,{'Content-Type':mime[path.extname(file)]||'application/octet-stream','Cache-Control':'no-store','X-Content-Type-Options':'nosniff'});res.end(await readFile(file))}catch{res.writeHead(500);res.end('Preview unavailable')}}).listen(Number(process.env.PORT || 4173),'127.0.0.1',()=>process.stdout.write('Forma local prototype: http://127.0.0.1:' + (process.env.PORT || 4173) + '\nPress Ctrl+C to stop.\n'));
