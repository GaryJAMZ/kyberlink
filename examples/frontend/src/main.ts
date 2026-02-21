// Copyright 2026 José Antonio Garibay Marcelo
// SPDX-License-Identifier: Apache-2.0

import './style.css'
import { KyberLinkClient } from './lib/kyberlink/client'

const client = new KyberLinkClient({
  gatewayUrl: 'http://localhost:45782'
});

document.querySelector<HTMLDivElement>('#app')!.innerHTML = `
  <div class="container">
    <header>
      <h1>KyberLink <span class="accent">Security Pro</span></h1>
      <p>3-Server Architecture Verification</p>
    </header>

    <div class="card">
      <div class="test-group">
        <div class="input-row">
          <input type="text" id="input-test1" placeholder="Enter custom message..." class="styled-input">
          <button id="btn-test1" class="btn">Send Text</button>
        </div>
        <div class="input-row">
          <input type="date" id="input-test2" class="styled-input">
          <button id="btn-test2" class="btn">Send Date</button>
        </div>
        <div class="input-row">
          <input type="password" id="input-test3" placeholder="Enter test password..." class="styled-input">
          <button id="btn-test3" class="btn">Send Password</button>
        </div>
      </div>

      <div class="test-group">
        <div class="log-header" style="background: transparent; border: none; padding-left: 0; color: white;">
          <span>Load Testing (Concurrency)</span>
        </div>
        <div class="input-row">
          <input type="number" id="input-load" placeholder="Number of requests (e.g. 50)" class="styled-input" min="1" max="5000">
          <button id="btn-load" class="btn" style="background: #e11d48;">RUN LOAD TEST</button>
        </div>
      </div>

      <div class="log-container">
        <div class="log-header">
          <span>Target Backend Console Simulation</span>
          <button id="clear-log" class="btn-sm">Clear</button>
        </div>
        <div id="log" class="log-content">
          <div class="log-entry system">Awaiting secure transmission...</div>
        </div>
      </div>
    </div>

    <footer>
      <p>Frontend > KyberLink Gateway > Final Go Backend</p>
    </footer>
  </div>
`

const logElement = document.querySelector<HTMLDivElement>('#log')!;

function addLog(msg: string, type: 'info' | 'success' | 'error' | 'system' = 'info') {
  const entry = document.createElement('div');
  entry.className = `log-entry ${type}`;
  const time = new Date().toLocaleTimeString();
  entry.innerHTML = `<span class="timestamp">[${time}]</span> ${msg}`;
  logElement.appendChild(entry);
  logElement.scrollTop = logElement.scrollHeight;
}

async function runLoadTest() {
  const inputEl = document.querySelector<HTMLInputElement>('#input-load')!;
  const count = parseInt(inputEl.value) || 10;
  if (count <= 0) return;

  addLog(`LOAD TEST: ${count} concurrent requests...`, 'system');

  const promises = [];
  const startTime = performance.now();
  let success = 0;
  let errors = 0;

  for (let i = 0; i < count; i++) {
    const p = client.send('/test1', 'POST', {
      type: 'load_test',
      id: i,
      timestamp: Date.now()
    }).then(() => {
      success++;
    }).catch((err: any) => {
      errors++;
      console.error(err);
    });
    promises.push(p);
  }

  await Promise.all(promises);
  const totalTime = performance.now() - startTime;

  addLog(`LOAD TEST COMPLETE`, 'system');
  addLog(`Total: ${totalTime.toFixed(0)}ms | Success: ${success} | Errors: ${errors}`, errors > 0 ? 'error' : 'success');
}

async function runTest(type: 'input' | 'date' | 'password') {
  let path = '';
  let payload: any = {};

  const inputEl1 = document.querySelector<HTMLInputElement>('#input-test1')!;
  const inputEl2 = document.querySelector<HTMLInputElement>('#input-test2')!;
  const inputEl3 = document.querySelector<HTMLInputElement>('#input-test3')!;

  switch (type) {
    case 'input':
      path = '/test1';
      payload = { type: 'form_input', value: inputEl1.value || 'Hello KyberLink!', user: 'Antonio' };
      break;
    case 'date':
      path = '/test2';
      const dateVal = inputEl2.value;
      payload = { type: 'system_date', value: dateVal ? new Date(dateVal).toISOString() : new Date().toISOString() };
      break;
    case 'password':
      path = '/test3';
      payload = { type: 'password_data', value: inputEl3.value || 'secret_PQC_2026', securityLevel: 'high' };
      break;
  }

  addLog(`Sending ${type} data to ${path}...`, 'info');
  try {
    const startTime = performance.now();
    const result = await client.send(path, 'POST', payload);
    const endTime = performance.now();

    addLog(`Response (${(endTime - startTime).toFixed(1)}ms):`, 'success');
    addLog(`<pre>${JSON.stringify(result.data, null, 2)}</pre>`, 'success');
  } catch (error: any) {
    addLog(`Error: ${error.message}`, 'error');
  }
}

document.querySelector('#btn-test1')?.addEventListener('click', () => runTest('input'));
document.querySelector('#btn-test2')?.addEventListener('click', () => runTest('date'));
document.querySelector('#btn-test3')?.addEventListener('click', () => runTest('password'));
document.querySelector('#btn-load')?.addEventListener('click', () => runLoadTest());
document.querySelector('#clear-log')?.addEventListener('click', () => {
  logElement.innerHTML = '<div class="log-entry system">Log cleared.</div>';
});
