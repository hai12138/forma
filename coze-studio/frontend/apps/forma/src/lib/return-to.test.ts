import { describe, expect, it } from 'vitest';

import { encodeReturnTo, safeReturnTo } from '../lib/return-to';

describe('safeReturnTo', () => {
  it('allows internal paths', () => {
    expect(safeReturnTo('/')).toBe('/');
    expect(safeReturnTo('/data/contracts/abc')).toBe('/data/contracts/abc');
    expect(safeReturnTo('/business?x=1')).toBe('/business?x=1');
  });

  it('denies open redirects', () => {
    expect(safeReturnTo('https://evil.example/')).toBe('/');
    expect(safeReturnTo('//evil.example')).toBe('/');
    expect(safeReturnTo('/\\evil')).toBe('/');
    expect(safeReturnTo('javascript:alert(1)')).toBe('/');
    expect(safeReturnTo(null)).toBe('/');
  });

  it('decodes encoded internal paths', () => {
    expect(safeReturnTo(encodeURIComponent('/data/health'))).toBe('/data/health');
  });

  it('encodeReturnTo is reversible for safe paths', () => {
    expect(safeReturnTo(decodeURIComponent(encodeReturnTo('/data')))).toBe('/data');
  });
});
