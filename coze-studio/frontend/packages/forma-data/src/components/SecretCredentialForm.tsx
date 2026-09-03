import { useState, type FormEvent } from 'react';

import type { FormaApiClient, FormaCredentialRef } from '@forma/api-client';

import { safeMutate } from '../utils/errors';

export interface SecretCredentialFormProps {
  client: FormaApiClient;
  canEdit: boolean;
  onCreated?: (cred: FormaCredentialRef) => void;
  onError?: (message: string) => void;
}

/**
 * Write-only credential form. On success the password field is cleared and
 * never echoed from the API response.
 */
export function SecretCredentialForm({
  client,
  canEdit,
  onCreated,
  onError,
}: SecretCredentialFormProps) {
  const [secretType, setSecretType] = useState('password');
  const [password, setPassword] = useState('');
  const [busy, setBusy] = useState(false);
  const [lastRefId, setLastRefId] = useState<string | null>(null);

  if (!canEdit) {
    return null;
  }

  const onSubmit = (e: FormEvent) => {
    e.preventDefault();
    if (busy) return;
    const submittedSecret = password;
    setBusy(true);
    void safeMutate(
      async () => {
        const resp = await client.createDataCredential({
          secret_type: secretType,
          secret: { password: submittedSecret },
        });
        setPassword('');
        setLastRefId(resp.data.credential_ref_id);
        onCreated?.(resp.data);
      },
      message => onError?.(message),
      [submittedSecret],
    ).finally(() => {
      setPassword('');
      setBusy(false);
    });
  };

  return (
    <form className="forma-panel" onSubmit={onSubmit} data-testid="secret-credential-form">
      <h3 style={{ marginTop: 0 }}>创建凭证</h3>
      <p className="forma-muted">密码仅用于提交，不会回显。</p>
      <div className="forma-form-row">
        <label htmlFor="cred-secret-type">密钥类型</label>
        <select
          id="cred-secret-type"
          value={secretType}
          onChange={ev => setSecretType(ev.target.value)}
        >
          <option value="password">password</option>
          <option value="headers">headers</option>
        </select>
      </div>
      <div className="forma-form-row">
        <label htmlFor="cred-password">密码</label>
        <input
          id="cred-password"
          type="password"
          autoComplete="new-password"
          value={password}
          onChange={ev => setPassword(ev.target.value)}
          data-testid="cred-password-input"
        />
      </div>
      <button className="forma-btn forma-btn-primary" type="submit" disabled={busy || !password}>
        {busy ? '提交中…' : '创建凭证'}
      </button>
      {lastRefId ? (
        <p className="forma-muted" data-testid="cred-created-ref">
          已创建：{lastRefId}
        </p>
      ) : null}
    </form>
  );
}
