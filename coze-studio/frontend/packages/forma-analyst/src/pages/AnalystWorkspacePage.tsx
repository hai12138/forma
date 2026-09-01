import { useCallback, useEffect, useMemo, useState } from 'react';
import { useNavigate } from 'react-router-dom';

import type {
  FormaApiClient,
  FormaAssertion,
  FormaAssertionEdit,
  FormaAnalystSession,
  FormaAnalystTurn,
  FormaBusiness,
  FormaConflict,
  FormaDiffResponse,
  FormaEvidence,
  FormaGap,
  FormaProposal,
  FormaProposalPreview,
  FormaTenant,
} from '@forma/api-client';
import { FormaApiError } from '@forma/api-client';

import '../styles/analyst.css';

export interface AnalystWorkspacePageProps {
  client: FormaApiClient;
  currentTenant: FormaTenant | null;
}

type SideTab = 'assertions' | 'evidence' | 'conflicts' | 'gaps' | 'proposal';

function clearAnalystWorkspaceState() {
  return {
    businessId: '',
    businesses: [] as FormaBusiness[],
    session: null as FormaAnalystSession | null,
    turns: [] as FormaAnalystTurn[],
    assertions: [] as FormaAssertion[],
    evidence: [] as FormaEvidence[],
    conflicts: [] as FormaConflict[],
    gaps: [] as FormaGap[],
    proposal: null as FormaProposal | null,
    proposalPreview: null as FormaProposalPreview | null,
    error: null as string | null,
    highlightEvidence: null as string | null,
    editTarget: null as FormaAssertion | null,
  };
}

export function AnalystWorkspacePage({ client, currentTenant }: AnalystWorkspacePageProps) {
  const navigate = useNavigate();
  const tenantId = currentTenant?.tenant_id ?? '';
  const [businesses, setBusinesses] = useState<FormaBusiness[]>([]);
  const [businessId, setBusinessId] = useState('');
  const [session, setSession] = useState<FormaAnalystSession | null>(null);
  const [turns, setTurns] = useState<FormaAnalystTurn[]>([]);
  const [assertions, setAssertions] = useState<FormaAssertion[]>([]);
  const [evidence, setEvidence] = useState<FormaEvidence[]>([]);
  const [conflicts, setConflicts] = useState<FormaConflict[]>([]);
  const [gaps, setGaps] = useState<FormaGap[]>([]);
  const [proposal, setProposal] = useState<FormaProposal | null>(null);
  const [proposalPreview, setProposalPreview] = useState<FormaProposalPreview | null>(null);
  const [tab, setTab] = useState<SideTab>('assertions');
  const [message, setMessage] = useState('');
  const [loading, setLoading] = useState(false);
  const [processing, setProcessing] = useState(false);
  const [error, setError] = useState<string | null>(null);
  const [highlightEvidence, setHighlightEvidence] = useState<string | null>(null);
  const [editTarget, setEditTarget] = useState<FormaAssertion | null>(null);
  const [editForm, setEditForm] = useState<FormaAssertionEdit>({
    assertion_type: '',
    subject_ref: '',
    predicate: '',
    object_value: '',
  });

  const assertionById = useMemo(() => {
    const map = new Map<string, FormaAssertion>();
    assertions.forEach(a => map.set(a.assertion_id, a));
    return map;
  }, [assertions]);

  const assertionsForEvidence = useCallback(
    (evidenceId: string) =>
      assertions.filter(a => a.evidence_ids.includes(evidenceId)),
    [assertions],
  );

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

  // Tenant switch: hard reset — never show prior tenant data.
  useEffect(() => {
    const cleared = clearAnalystWorkspaceState();
    setBusinesses(cleared.businesses);
    setBusinessId(cleared.businessId);
    setSession(cleared.session);
    setTurns(cleared.turns);
    setAssertions(cleared.assertions);
    setEvidence(cleared.evidence);
    setConflicts(cleared.conflicts);
    setGaps(cleared.gaps);
    setProposal(cleared.proposal);
    setProposalPreview(cleared.proposalPreview);
    setError(cleared.error);
    setHighlightEvidence(cleared.highlightEvidence);
    setEditTarget(cleared.editTarget);

    if (!tenantId) return;
    void (async () => {
      try {
        const resp = await client.listBusinesses();
        const list = resp.data ?? [];
        setBusinesses(list);
        if (list.length) {
          setBusinessId(list[0].business_id);
        }
      } catch (err) {
        setError(err instanceof Error ? err.message : '加载业务失败');
      }
    })();
  }, [tenantId, client]);

  useEffect(() => {
    if (businessId) {
      void refreshAnalystData();
    }
  }, [businessId, refreshAnalystData]);

  useEffect(() => {
    void loadSessionTurns();
  }, [loadSessionTurns]);

  useEffect(() => {
    if (!businessId || !proposal) {
      setProposalPreview(null);
      return;
    }
    void (async () => {
      try {
        const resp = await client.getProposalPreview(businessId, proposal.proposal_id);
        setProposalPreview(resp.data ?? null);
      } catch {
        setProposalPreview(null);
      }
    })();
  }, [businessId, proposal, client]);

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
      setProposalPreview(null);
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

  const retryAnalysis = async (turnId: string) => {
    if (!businessId || !session) return;
    setProcessing(true);
    setError(null);
    try {
      const resp = await client.retryAnalystTurnAnalysis(businessId, session.session_id, turnId);
      await loadSessionTurns();
      await refreshAnalystData();
      if (resp.data?.model_failed) {
        setError(resp.data.model_error || '重试分析失败');
      } else {
        setError(null);
      }
    } catch (err) {
      setError(err instanceof Error ? err.message : '重试失败');
    } finally {
      setProcessing(false);
    }
  };

  const confirmAssertion = async (id: string, edit?: FormaAssertionEdit) => {
    if (!businessId) return;
    try {
      await client.confirmAssertion(businessId, id, {
        comment: edit ? 'edit and confirm via analyst UI' : 'confirmed via analyst UI',
        edit,
      });
      setEditTarget(null);
      await refreshAnalystData();
    } catch (err) {
      setError(err instanceof Error ? err.message : '确认失败');
    }
  };

  const openEditConfirm = (a: FormaAssertion) => {
    setEditTarget(a);
    setEditForm({
      assertion_type: a.assertion_type,
      subject_ref: a.subject_ref,
      predicate: a.predicate,
      object_value: a.object_value,
    });
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
      setProposal(prev => (prev ? { ...prev, status: 'APPLIED' } : prev));
      if (resp.data?.revision_no) {
        navigate(`/business/${businessId}`);
      }
    } catch (err) {
      if (err instanceof FormaApiError && err.errorKey === 'FORMA_PROPOSAL_STALE') {
        setError('提案已过期，请重新生成');
        setProposal(prev => (prev ? { ...prev, status: 'STALE' } : prev));
      } else {
        setError(err instanceof Error ? err.message : '应用失败');
      }
    } finally {
      setLoading(false);
    }
  };

  const focusEvidence = (evidenceId: string) => {
    setTab('evidence');
    setHighlightEvidence(evidenceId);
  };

  const focusAssertion = (assertionId: string) => {
    setTab('assertions');
    const a = assertionById.get(assertionId);
    if (a?.evidence_ids[0]) {
      setHighlightEvidence(a.evidence_ids[0]);
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

  const askGap = async (gap: FormaGap) => {
    if (!businessId || !session) return;
    setProcessing(true);
    setError(null);
    try {
      const resp = await client.askAnalystGap(businessId, session.session_id, gap.gap_id);
      if (resp.data?.analyst_turn) {
        setTurns(prev => {
          const exists = prev.some(t => t.turn_id === resp.data!.analyst_turn.turn_id);
          if (exists) {
            return prev;
          }
          return [...prev, resp.data!.analyst_turn].sort((a, b) => a.sequence - b.sequence);
        });
      }
      await loadSessionTurns();
    } catch (err) {
      setError(err instanceof Error ? err.message : 'Gap ask failed');
    } finally {
      setProcessing(false);
    }
  };

  const canApplyProposal =
    proposal?.status === 'READY_FOR_REVIEW' &&
    proposalPreview?.validation_valid &&
    proposalPreview.current_revision === proposal.base_revision;

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
              setProposalPreview(null);
              setAssertions([]);
              setEvidence([]);
              setConflicts([]);
              setGaps([]);
            }}
            style={{ marginLeft: 8 }}
            data-testid="business-select"
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

      {!tenantId && (
        <p className="forma-placeholder" data-testid="tenant-empty">请先选择租户</p>
      )}
      {tenantId && businesses.length === 0 && (
        <p className="forma-placeholder" data-testid="business-empty">当前租户暂无业务资产</p>
      )}

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
              {(t.analysis_status === 'FAILED' ||
                t.analysis_status === 'EXTRACTION_FAILED' ||
                t.analysis_status === 'RESPONSE_FAILED') &&
                t.speaker === 'USER' && (
                <div className="forma-analyst-processing">
                  分析失败 — 您的输入已保存
                  <button
                    type="button"
                    className="forma-btn"
                    style={{ marginLeft: 8 }}
                    data-testid="retry-analysis"
                    disabled={processing}
                    onClick={() => void retryAnalysis(t.turn_id)}
                  >
                    Retry Analysis
                  </button>
                </div>
              )}
            </div>
          ))}
          {!session && businesses.length > 0 && (
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
                  onClick={() => {
                    if (a.evidence_ids[0]) focusEvidence(a.evidence_ids[0]);
                  }}
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
                  {a.source_marker === 'MANUAL_MODIFIED' && (
                    <div style={{ fontSize: 11, color: '#059669' }}>人工编辑后确认</div>
                  )}
                  {a.evidence_ids.length > 0 && (
                    <div style={{ fontSize: 11, color: '#9ca3af' }}>
                      Evidence: {a.evidence_ids.map(id => id.slice(0, 8)).join(', ')}
                    </div>
                  )}
                  {a.status === 'PROPOSED' && (
                    <div className="forma-analyst-card-actions">
                      <button
                        type="button"
                        className="forma-btn forma-btn-primary"
                        data-testid="confirm-assertion"
                        onClick={e => {
                          e.stopPropagation();
                          void confirmAssertion(a.assertion_id);
                        }}
                      >
                        Confirm
                      </button>
                      <button
                        type="button"
                        className="forma-btn"
                        data-testid="edit-confirm-assertion"
                        onClick={e => {
                          e.stopPropagation();
                          openEditConfirm(a);
                        }}
                      >
                        Edit & Confirm
                      </button>
                      <button
                        type="button"
                        className="forma-btn"
                        onClick={e => {
                          e.stopPropagation();
                          void rejectAssertion(a.assertion_id);
                        }}
                      >
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
                  data-testid="evidence-card"
                  style={{
                    outline: highlightEvidence === e.evidence_id ? '2px solid #1a56db' : undefined,
                  }}
                >
                  <div className="forma-analyst-badge">{e.source_type}</div>
                  <div>{e.quote}</div>
                  <div style={{ fontSize: 11, color: '#9ca3af' }}>turn: {e.turn_id.slice(0, 12)}…</div>
                  <div style={{ marginTop: 6 }}>
                    <strong style={{ fontSize: 12 }}>Linked Assertions</strong>
                    {assertionsForEvidence(e.evidence_id).map(a => (
                      <button
                        key={a.assertion_id}
                        type="button"
                        className="forma-btn"
                        style={{ margin: '4px 4px 0 0', fontSize: 11 }}
                        onClick={() => focusAssertion(a.assertion_id)}
                      >
                        {a.object_value} ({a.status})
                      </button>
                    ))}
                    {!assertionsForEvidence(e.evidence_id).length && (
                      <span className="forma-placeholder" style={{ fontSize: 11 }}>无关联事实</span>
                    )}
                  </div>
                </div>
              ))}
            </div>
          )}

          {tab === 'conflicts' && (
            <div data-testid="conflicts-panel">
              {conflicts.map(c => {
                const aA = assertionById.get(c.assertion_id_a);
                const aB = assertionById.get(c.assertion_id_b);
                return (
                  <div key={c.conflict_id} className="forma-analyst-card" data-testid="conflict-card">
                    <div>
                      <span className="forma-analyst-badge">{c.status}</span>
                      {c.subject_ref} · {c.predicate}
                    </div>
                    <div className="forma-analyst-conflict-vs">
                      <div>
                        <strong>事实 A</strong>
                        <div>{aA?.object_value ?? c.assertion_id_a}</div>
                        {aA?.evidence_ids[0] && (
                          <button type="button" className="forma-btn" onClick={() => focusEvidence(aA.evidence_ids[0])}>
                            查看证据
                          </button>
                        )}
                        {aA?.status === 'PROPOSED' && (
                          <button
                            type="button"
                            className="forma-btn forma-btn-primary"
                            onClick={() => void confirmAssertion(aA.assertion_id)}
                          >
                            Confirm A
                          </button>
                        )}
                      </div>
                      <div>VS</div>
                      <div>
                        <strong>事实 B</strong>
                        <div>{aB?.object_value ?? c.assertion_id_b}</div>
                        {aB?.evidence_ids[0] && (
                          <button type="button" className="forma-btn" onClick={() => focusEvidence(aB.evidence_ids[0])}>
                            查看证据
                          </button>
                        )}
                        {aB?.status === 'PROPOSED' && (
                          <button
                            type="button"
                            className="forma-btn"
                            onClick={() => void rejectAssertion(aB.assertion_id)}
                          >
                            Reject B
                          </button>
                        )}
                      </div>
                    </div>
                  </div>
                );
              })}
            </div>
          )}

          {tab === 'gaps' && (
            <div data-testid="gaps-panel">
              {gaps.map(g => (
                <div key={g.gap_id} className="forma-analyst-card">
                  <div>{g.question}</div>
                  <button type="button" className="forma-btn" data-testid="gap-ask" onClick={() => void askGap(g)} disabled={processing}>
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
                  <div>Current revision r{proposalPreview?.current_revision ?? '?'}</div>
                  <div>Assertions used: {proposalPreview?.assertion_count ?? proposal.assertion_ids.length}</div>
                  <div>
                    Validation:{' '}
                    {proposalPreview?.validation_valid
                      ? 'VALID'
                      : proposalPreview?.validation_error ?? 'checking…'}
                  </div>
                  <div>Status: {proposal.status}</div>
                  {proposalPreview?.diff && proposalPreview?.impact && (
                    <ProposalDiffView
                      data={{
                        diff: proposalPreview.diff,
                        impact: proposalPreview.impact,
                      }}
                    />
                  )}
                  <button
                    type="button"
                    className="forma-btn forma-btn-primary"
                    data-testid="apply-proposal"
                    disabled={!canApplyProposal || loading}
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

      {editTarget && (
        <div className="forma-analyst-modal" data-testid="edit-confirm-modal">
          <div className="forma-analyst-card">
            <h4>Edit & Confirm</h4>
            <p style={{ fontSize: 12, color: '#6b7280' }}>
              AI 原始识别: {editTarget.object_value} ({editTarget.assertion_type})
            </p>
            <label>
              Type
              <input
                value={editForm.assertion_type}
                onChange={e => setEditForm(f => ({ ...f, assertion_type: e.target.value }))}
              />
            </label>
            <label>
              Subject
              <input
                value={editForm.subject_ref}
                onChange={e => setEditForm(f => ({ ...f, subject_ref: e.target.value }))}
              />
            </label>
            <label>
              Predicate
              <input
                value={editForm.predicate}
                onChange={e => setEditForm(f => ({ ...f, predicate: e.target.value }))}
              />
            </label>
            <label>
              Object Value
              <input
                value={editForm.object_value}
                onChange={e => setEditForm(f => ({ ...f, object_value: e.target.value }))}
              />
            </label>
            <div className="forma-analyst-card-actions">
              <button
                type="button"
                className="forma-btn forma-btn-primary"
                onClick={() => void confirmAssertion(editTarget.assertion_id, editForm)}
              >
                确认后的业务事实
              </button>
              <button type="button" className="forma-btn" onClick={() => setEditTarget(null)}>
                取消
              </button>
            </div>
          </div>
        </div>
      )}
    </div>
  );
}

function ProposalDiffView({ data }: { data: FormaDiffResponse }) {
  const sections = [
    ['Nodes', data.diff.nodes],
    ['Edges', data.diff.edges],
    ['Rules', data.diff.rules],
    ['States', data.diff.states],
  ] as const;

  return (
    <div data-testid="proposal-semantic-diff">
      <p className="forma-placeholder">
        r{data.diff.from_revision} → r{data.diff.to_revision}
        {data.impact.semantic_changed ? ' · semantic changed' : ' · no semantic change'}
      </p>
      {sections.map(([title, block]) => (
        <div key={title} className="forma-biz-diff-block">
          <b>{title}</b>
          <ul>
            {block.added.map(id => (
              <li key={`a-${id}`} className="forma-biz-diff-added">Added: {id}</li>
            ))}
            {block.removed.map(id => (
              <li key={`r-${id}`} className="forma-biz-diff-removed">Removed: {id}</li>
            ))}
            {block.modified.map(id => (
              <li key={`m-${id}`} className="forma-biz-diff-modified">Modified: {id}</li>
            ))}
            {!block.added.length && !block.removed.length && !block.modified.length && (
              <li className="forma-placeholder">无变更</li>
            )}
          </ul>
        </div>
      ))}
    </div>
  );
}
