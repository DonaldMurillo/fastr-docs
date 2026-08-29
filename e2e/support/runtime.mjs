import fs from 'node:fs';
import path from 'node:path';
import { fileURLToPath } from 'node:url';

const here = path.dirname(fileURLToPath(import.meta.url));
const runtimePath = path.resolve(here, '..', '.runtime.json');

export const runtime = () => JSON.parse(fs.readFileSync(runtimePath, 'utf8'));
