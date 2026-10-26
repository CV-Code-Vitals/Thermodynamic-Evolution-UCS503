import { readFileSync } from 'node:fs';

const script = readFileSync(0, 'utf8');
const encoded = Buffer.from(script, 'utf8').toString('base64');
process.stdout.write(`printf '%s' '${encoded}' | base64 -d | bash`);
