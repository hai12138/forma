import { useCallback, useEffect, useMemo, useState } from 'react';
import { useNavigate } from 'react-router-dom';

import type {
  FormaApiClient,
  FormaAssertion,
  FormaAnalystSession,
  FormaAnalystTurn,
  FormaBusiness,
  FormaConflict,
  FormaEvidence,
  FormaGap,
  FormaProposal,
  FormaTenant,
} from '@forma/api-client';
import { FormaApiError } from '@forma/api-client';

import '../styles/analyst.css';

export interface AnalystWorkspacePageProps {
  client: FormaApiClient;
  currentTenant: FormaTenant | null;
}

type SideTab = 'assertions' | 'evidence' | 'conflicts' | 'gaps' | 'proposal';

export function AnalystWorkspacePage({ client, currentTenant }: AnalystWorkspacePageProps) {
  const navigate = useNavigate();
  const [businesses, setBusinesses] = useState<FormaBusiness[]>([]);
  const [businessId, setBusinessId] = useState('');
  const [session, setSession] = useState<FormaAnalystSession | null>(null);
  const [turns, setTurns] = useState<FormaAnalystTurn[]>([]);
  const [assertions, setAssertions] = useState<FormaAssertion[]>([]);
  const [evidence, setEvidence] = useState<FormaEvidence[]>([]);
  const [conflicts, setConflicts] = useState<FormaConflict[]>([]);
  const [gaps, setGaps] = useState<FormaGap[]>([]);
  const [proposal, setProposal] = useState<FormaProposal | null>(null);
  const [tab, setTab] = useState<SideTab>('assertions');
  const [message, setMessage] = useState('');
  const [loading, setLoading] = useState(false);
  const [processing, setProcessing] = useState(false);
  const [error, setError] = useState<string | null>(null);
  const [highlightEvidence, setHighlightEvidence] = useState<string | null>(null);

  const loadBusinesses = useCallback(async () => {
    try {
      const resp = await client.listBusinesses();
      setBusinesses(resp.data ?? []);
      if (!businessId && resp.data?.length) {
        setBusinessId(resp.data[0].business_id);
      }
    } catch (err) {
      setError(err instanceof Error ? err.message : '加载业务失败');
    }
  }, [client, businessId]);

  const refreshAnalystData = useCallback(async () => {
    if (!businessId) return;
    const [a, e, c, g] = await Promise.all([
      client.listAssertions(businessId),
      client.listEvidence(businessId),
      client.listConflicts(businessId),
      client.listGaps(businessId),
    ]);
    setAssertions(a.data ?? []);
    setEvidence(e.data ?? []);
    setConflicts(c.data ?? []);
    setGaps(g.data ?? []);
  }, [client, businessId]);

  const loadSessionTurns = useCallback(async () => {
    if (!businessId || !session) return;
    const resp = await client.listAnalystTurns(businessId, session.session_id);
    setTurns(resp.data ?? []);
  }, [client, businessId, session]);

  useEffect(() => {
    void loadBusinesses();
  }, [loadBusinesses, currentTenant?.tenant_id]);

  useEffect(() => {
    if (businessId) {
      void refreshAnalystData();
    }
  }, [businessId, refreshAnalystData]);

  useEffect(() => {
    void loadSessionTurns();
  }, [loadSessionTurns]);

  const startSession = async () => {
    if (!businessId) return;
    setLoading(true);
    setError(null);
    try {
      const resp = await client.createAnalystSession(businessId, {
        title: '业务访谈',
        confirmation_policy: 'DEVELOPMENT',
      });
      setSession(resp.data);
      setTurns([]);
      setProposal(null);
    } catch (err) {
      setError(err instanceof Error ? err.message : '创建会话失败');
    } finally {
      setLoading(false);
    }
  };

  const submitTurn = async () => {
    if (!businessId || !session || !message.trim()) return;
    setProcessing(true);
    setError(null);
    const clientRequestId = `cr_${Date.now()}`;
    try {
      const resp = await client.submitAnalystTurn(businessId, session.session_id, {
        content: message.trim(),
        client_request_id: clientRequestId,
      });
      setMessage('');
      if (resp.data?.user_turn) {
        setTurns(prev => {
          const next = [...prev];
          if (resp.data?.analyst_turn) {
            next.push(resp.data.user_turn, resp.data.analyst_turn);
          } else {
            next.push(resp.data!.user_turn);
          }
          return next;
        });
      } else {
        await loadSessionTurns();
      }
      await refreshAnalystData();
      if (resp.data?.model_failed) {
        setError(resp.data.model_error || '模型分析失败，但您的输入已保存');
      }
    } catch (err) {
      setError(err instanceof Error ? err.message : '提交失败');
    } finally {
      setProcessing(false);
    }
  };

  const confirmAssertion = async (id: string) => {
    if (!businessId) return;
    try {
      await client.confirmAssertion(businessId, id, { comment: 'confirmed via analyst UI' });
      await refreshAnalystData();
    } catch (err) {
      setError(err instanceof Error ? err.message : '确认失败');
    }
  };

  const rejectAssertion = async (id: string) => {
    if (!businessId) return;
    try {
      await client.rejectAssertion(businessId, id, { comment: 'rejected via analyst UI' });
      await refreshAnalystData();
    } catch (err) {
      setError(err instanceof Error ? err.message : '拒绝失败');
    }
  };

  const generateProposal = async () => {
    if (!businessId || !session) return;
    setLoading(true);
    try {
      const resp = await client.createProposal(businessId, { session_id: session.session_id });
      setProposal(resp.data);
      setTab('proposal');
    } catch (err) {
      setError(err instanceof Error ? err.message : '生成提案失败');
    } finally {
      setLoading(false);
    }
  };

  const applyProposal = async () => {
    if (!businessId || !proposal) return;
    setLoading(true);
    try {
      const resp = await client.applyProposal(businessId, proposal.proposal_id);
      await refreshAnalystData();
      setProposal(prev =>
        prev ? { ...prev, status: 'APPLIED' } : prev,
      );
      if (resp.data?.revision_no) {
        navigate(`/business/${businessId}`);
      }
    } catch (err) {
      if (err instanceof FormaApiError && err.errorKey === 'FORMA_PROPOSAL_STALE') {
        setError('提案已过期，请重新生成');
      } else {
        setError(err instanceof Error ? err.message : '应用失败');
      }
    } finally {
      setLoading(false);
    }
  };

  const proposedCount = useMemo(
    () => assertions.filter(a => a.status === 'PROPOSED').length,
    [assertions],
  );
  const confirmedCount = useMemo(
    () => assertions.filter(a => a.status === 'CONFIRMED').length,
    [assertions],
  );

  const askGap = (gap: FormaGap) => {
    setMessage(gap.question);
  };

  return (
    <div className="forma-panel">
      <h1 style={{ marginTop: 0 }}>AI 业务分析师</h1>
      <p className="forma-placeholder" style={{ marginBottom: 16 }}>
        访谈 → Evidence → Assertion → Confirmation → Proposal → Apply。AI 不会直接修改业务模型。
      </p>

      {error && (
        <div className="forma-analyst-card" style={{ borderColor: '#fca5a5', color: '#b91c1c' }}>
          {error}
        </div>
      )}

      <div style={{ marginBottom: 12 }}>
        <label>
          业务资产：
          <select
            value={businessId}
            onChange={e => {
              setBusinessId(e.target.value);
              setSession(null);
              setTurns([]);
              setProposal(null);
            }}
            style={{ marginLeft: 8 }}
          >
            <option value="">选择业务</option>
            {businesses.map(b => (
              <option key={b.business_id} value={b.business_id}>
                {b.name} (r{b.current_revision})
              </option>
            ))}
          </select>
        </label>
        <button
          type="button"
          className="forma-btn"
          style={{ marginLeft: 12 }}
          disabled={!businessId || loading}
          onClick={() => void startSession()}
        >
          开始访谈
        </button>
        {session && (
          <span className="forma-analyst-badge" style={{ marginLeft: 12 }}>
            会话 {session.session_id.slice(0, 12)}…
          </span>
        )}
      </div>

      <div className="forma-analyst-layout">
        <div className="forma-analyst-thread">
          <h3>访谈</h3>
          {turns.map(t => (
            <div
              key={t.turn_id}
              className={`forma-analyst-turn ${t.speaker === 'USER' ? 'user' : 'analyst'}`}
              data-testid={`turn-${t.speaker.toLowerCase()}`}
            >
              <strong>{t.speaker === 'USER' ? '用户' : '分析师'}</strong>
              <div>{t.content}</div>
              {t.analysis_status === 'FAILED' && (
                <div className="forma-analyst-processing">分析失败 — 可重试或继续输入</div>
              )}
            </div>
          ))}
          {!session && (
            <p className="forma-placeholder">选择业务并点击「开始访谈」</p>
          )}
          {processing && <div className="forma-analyst-processing">正在分析业务事实…</div>}
          {session && (
            <div className="forma-analyst-input-row">
              <textarea
                value={message}
                onChange={e => setMessage(e.target.value)}
                placeholder="描述业务流程，例如：员工报修、维修人员处理、管理员关闭工单…"
                data-testid="analyst-input"
              />
              <button
                type="button"
                className="forma-btn forma-btn-primary"
                disabled={processing || !message.trim()}
                data-testid="analyst-submit"
                onClick={() => void submitTurn()}
              >
                发送
              </button>
            </div>
          )}
          {proposedCount > 0 && (
            <div className="forma-analyst-badge" data-testid="assertion-count">
              识别到 {proposedCount} 条待确认业务事实
            </div>
          )}
        </div>

        <div className="forma-analyst-side">
          <div className="forma-analyst-tabs">
            {(['assertions', 'evidence', 'conflicts', 'gaps', 'proposal'] as SideTab[]).map(t => (
              <button
                key={t}
                type="button"
                className={`forma-analyst-tab ${tab === t ? 'active' : ''}`}
                onClick={() => setTab(t)}
              >
                {t === 'assertions' && `Assertions (${assertions.length})`}
                {t === 'evidence' && `Evidence (${evidence.length})`}
                {t === 'conflicts' && `Conflicts (${conflicts.length})`}
                {t === 'gaps' && `Gaps (${gaps.length})`}
                {t === 'proposal' && 'Proposal'}
              </button>
            ))}
          </div>

          {tab === 'assertions' && (
            <div data-testid="assertions-panel">
              {assertions.map(a => (
                <div
                  key={a.assertion_id}
                  className="forma-analyst-card"
                  data-testid="assertion-card"
                  onMouseEnter={() => setHighlightEvidence(a.evidence_ids[0] ?? null)}
                >
                  <div>
                    <span className="forma-analyst-badge">{a.assertion_type}</span>
                    <strong style={{ marginLeft: 6 }}>{a.object_value}</strong>
                  </div>
                  <div>confidence: {a.confidence.toFixed(2)} · {a.status}</div>
                  <div style={{ fontSize: 12, color: '#6b7280' }}>
                    {a.subject_ref} · {a.predicate}
                  </div>
                  {a.status === 'PROPOSED' && (
                    <div className="forma-analyst-card-actions">
                      <button type="button" className="forma-btn forma-btn-primary" data-testid="confirm-assertion" onClick={() => void confirmAssertion(a.assertion_id)}>
                        Confirm
                      </button>
                      <button type="button" className="forma-btn" onClick={() => void rejectAssertion(a.assertion_id)}>
                        Reject
                      </button>
                    </div>
                  )}
                </div>
              ))}
            </div>
          )}

          {tab === 'evidence' && (
            <div data-testid="evidence-panel">
              {evidence.map(e => (
                <div
                  key={e.evidence_id}
                  className="forma-analyst-card"
                  style={{
                    outline: highlightEvidence === e.evidence_id ? '2px solid #1a56db' : undefined,
                  }}
                >
                  <div className="forma-analyst-badge">{e.source_type}</div>
                  <div>{e.quote}</div>
                  <div style={{ fontSize: 11, color: '#9ca3af' }}>turn: {e.turn_id}</div>
                </div>
              ))}
            </div>
          )}

          {tab === 'conflicts' && (
            <div data-testid="conflicts-panel">
              {conflicts.map(c => (
                <div key={c.conflict_id} className="forma-analyst-card">
                  <div>{c.subject_ref} · {c.predicate}</div>
                  <div>A: {c.assertion_id_a}</div>
                  <div>B: {c.assertion_id_b}</div>
                </div>
              ))}
            </div>
          )}

          {tab === 'gaps' && (
            <div data-testid="gaps-panel">
              {gaps.map(g => (
                <div key={g.gap_id} className="forma-analyst-card">
                  <div>{g.question}</div>
                  <button type="button" className="forma-btn" onClick={() => askGap(g)}>
                    Ask This
                  </button>
                </div>
              ))}
            </div>
          )}

          {tab === 'proposal' && (
            <div data-testid="proposal-panel">
              <button
                type="button"
                className="forma-btn forma-btn-primary"
                disabled={confirmedCount === 0 || loading}
                data-testid="generate-proposal"
                onClick={() => void generateProposal()}
              >
                Generate Business Model Proposal
              </button>
              {proposal && (
                <div className="forma-analyst-proposal-preview">
                  <div>Base revision r{proposal.base_revision}</div>
                  <div>Operations: {proposal.patch?.operations?.length ?? 0}</div>
                  <div>Status: {proposal.status}</div>
                  <ul>
                    {proposal.patch?.operations?.map((op, i) => (
                      <li key={i}>{op.op} {op.node?.name ?? op.rule?.name ?? op.state?.name ?? op.target_id}</li>
                    ))}
                  </ul>
                  <button
                    type="button"
                    className="forma-btn forma-btn-primary"
                    data-testid="apply-proposal"
                    disabled={proposal.status !== 'READY_FOR_REVIEW' || loading}
                    onClick={() => void applyProposal()}
                  >
                    Apply to Business Model
                  </button>
                </div>
              )}
            </div>
          )}
        </div>
      </div>
    </div>
  );
}
