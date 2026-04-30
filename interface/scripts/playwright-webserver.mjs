import fs from 'node:fs';
import path from 'node:path';
import process from 'node:process';
import { spawn } from 'node:child_process';

const interfaceRoot = path.resolve(import.meta.dirname, '..');
const standaloneServerCandidates = [
    path.join(interfaceRoot, '.next', 'standalone', 'server.js'),
    path.join(interfaceRoot, '.next', 'standalone', 'interface', 'server.js'),
];
const standaloneServer = standaloneServerCandidates.find((candidate) => fs.existsSync(candidate));
const standaloneRoot = standaloneServer ? path.dirname(standaloneServer) : path.join(interfaceRoot, '.next', 'standalone');
const standaloneStatic = path.join(standaloneRoot, '.next', 'static');
const rootStatic = path.join(interfaceRoot, '.next', 'static');
const nextStart = path.join(interfaceRoot, 'node_modules', 'next', 'dist', 'bin', 'next');
const bindHost =
    process.env.INTERFACE_BIND_HOST ??
    process.env.MYCELIS_INTERFACE_BIND_HOST ??
    process.env.MYCELIS_INTERFACE_HOST ??
    '127.0.0.1';
const port = process.env.PLAYWRIGHT_PORT ?? process.env.INTERFACE_PORT ?? '3100';

const childEnv = {
    ...process.env,
    HOSTNAME: bindHost,
    PORT: port,
};

if (!fs.existsSync(rootStatic) && !fs.existsSync(standaloneStatic)) {
    console.error('Playwright webserver cannot start: missing .next static assets. Run npm run build again.');
    process.exit(1);
}

if (standaloneServer && fs.existsSync(rootStatic) && !fs.existsSync(standaloneStatic)) {
    fs.mkdirSync(path.dirname(standaloneStatic), { recursive: true });
    fs.cpSync(rootStatic, standaloneStatic, { recursive: true });
}

const command = standaloneServer && fs.existsSync(standaloneStatic)
    ? ['node', standaloneServer]
    : ['node', nextStart, 'start', '--hostname', bindHost, '--port', port];

const child = spawn(command[0], command.slice(1), {
    cwd: interfaceRoot,
    env: childEnv,
    stdio: 'inherit',
});

child.on('exit', (code, signal) => {
    if (signal) {
        process.kill(process.pid, signal);
        return;
    }
    process.exit(code ?? 0);
});
