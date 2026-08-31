import { useRef, useState } from 'react';
import {
  Plus,
  Search,
  Copy,
  Download,
  Upload,
  Bot,
  Trash2,
  Play,
  RotateCcw,
  Save,
} from 'lucide-react';
import { Button } from '@/components/ui/button';
import { Input } from '@/components/ui/input';
import { Textarea } from '@/components/ui/textarea';
import { Checkbox } from '@/components/ui/checkbox';
import {
  Heading,
  Panel,
  Tabs,
  Badge,
  Notice,
  Rows,
  Field,
  Modal,
  Empty,
} from '../shared';
import {
  type Agent,
  capabilities,
  validateImport,
  download,
} from '@/lib/domain';
import { useStore } from '@/lib/store';
export default function Agents() {
  const { state, update, notify } = useStore();
  const [query, setQuery] = useState('');
  const [filter, setFilter] = useState('全部 Agent');
  const [draft, setDraft] = useState<Agent | null>(null);
  const [deleteId, setDeleteId] = useState('');
  const [tab, setTab] = useState('业务定义');
  const [test, setTest] = useState('A 座电梯故障，请紧急派单。');
  const [result, setResult] = useState('');
  const [busy, setBusy] = useState(false);
  const [error, setError] = useState('');
  const file = useRef<HTMLInputElement>(null);
  const visible = state.agents.filter(
    (a) =>
      (filter === '回收站' ? a.deleted : !a.deleted) &&
      (filter === '草稿' ? a.status === '草稿' : true) &&
      (a.name + a.role + a.description)
        .toLowerCase()
        .includes(query.toLowerCase()),
  );
  const create = () => {
    setTab('业务定义');
    setResult('');
    setError('');
    setDraft({
      id: crypto.randomUUID(),
      name: '',
      description: '',
      role: '',
      version: '0.1.0',
      status: '草稿',
      capabilities: [],
      knowledge: '园区服务手册',
      permission: 'tenant:current · 最小权限',
      interaction: '写操作前确认；缺少必要信息时澄清；失败时转人工。',
      deleted: false,
      history: [],
    });
  };
  const save = () => {
    if (!draft) return;
    if (
      !draft.name.trim() ||
      !draft.role.trim() ||
      draft.capabilities.length === 0
    ) {
      setError('请输入名称、业务 Role，并至少选择一个能力。');
      return;
    }
    const exists = state.agents.some((a) => a.id === draft.id);
    update(
      {
        agents: exists
          ? state.agents.map((a) =>
              a.id === draft.id ? { ...draft, status: '草稿' } : a,
            )
          : [...state.agents, draft],
      },
      exists ? 'Agent 草稿已保存，相关评测已失效' : '新 Agent 已创建',
      true,
    );
    setDraft(null);
  };
  return (
    <>
      <Heading
        eyebrow="AGENT ASSET / BUSINESS AGENT CENTER"
        title="每一个 Agent，都专注一项业务。"
        description="业务角色、上下文、能力、规则、权限与交互策略。通用对话能力由平台 Kernel 统一提供。"
      >
        <Button variant="outline" onClick={() => file.current?.click()}>
          <Upload size={15} />
          导入
        </Button>
        <Button onClick={create}>
          <Plus size={15} />
          创建 Agent
        </Button>
      </Heading>
      <input
        ref={file}
        type="file"
        accept="application/json,.json"
        hidden
        onChange={async (e) => {
          const f = e.target.files?.[0];
          if (!f) return;
          try {
            if (f.size > 1000000) throw Error('文件不得超过 1 MB');
            const agents = validateImport(JSON.parse(await f.text()));
            update(
              { agents: [...state.agents, ...agents] },
              '已导入 ' + agents.length + ' 个 Agent，均为待验证草稿',
              true,
            );
          } catch (err) {
            notify('导入失败：' + (err as Error).message);
          }
          e.target.value = '';
        }}
      />
      <div className="toolbar">
        <Tabs
          items={['全部 Agent', '草稿', '回收站']}
          value={filter}
          onChange={setFilter}
        />
        <div className="search-field">
          <Search size={16} />
          <Input
            placeholder="搜索 Agent、角色或业务…"
            aria-label="搜索 Agent"
            value={query}
            onChange={(e) => setQuery(e.target.value)}
          />
        </div>
        <Button
          variant="outline"
          onClick={() =>
            download(
              'business-agents.json',
              state.agents.filter((a) => !a.deleted),
            )
          }
        >
          <Download size={14} />
          导出全部
        </Button>
      </div>
      {!visible.length ? (
        <Panel>
          <Empty
            title="没有匹配的 Agent"
            description="更换搜索条件，或创建新的业务 Agent。"
            action="创建 Agent"
            onAction={create}
          />
        </Panel>
      ) : (
        <div className="agent-grid">
          {visible.map((a, i) => (
            <article className="agent-card" key={a.id}>
              <div className="agent-card-top">
                <span
                  className={
                    'icon-tile ' + ['blue', 'violet', 'teal', 'orange'][i % 4]
                  }
                >
                  <Bot size={23} />
                </span>
                <Badge
                  tone={
                    a.deleted
                      ? 'danger'
                      : a.status === '已发布'
                        ? 'success'
                        : 'neutral'
                  }
                >
                  {a.deleted ? '已删除' : a.status}
                </Badge>
              </div>
              <button
                className="agent-title"
                onClick={() => {
                  setDraft(structuredClone(a));
                  setTab('业务定义');
                  setResult('');
                  setError('');
                }}
              >
                <h2>{a.name}</h2>
                <p>{a.description}</p>
              </button>
              <div className="agent-metadata">
                <span>{a.capabilities.length} 项能力</span>
                <span>v{a.version}</span>
                <span>企业级权限</span>
              </div>
              <div className="agent-owner">
                <span className="avatar small">{a.name[0]}</span>
                {a.role}
              </div>
              <div className="agent-card-footer">
                {a.deleted ? (
                  <Button
                    variant="outline"
                    onClick={() =>
                      update(
                        {
                          agents: state.agents.map((x) =>
                            x.id === a.id ? { ...x, deleted: false } : x,
                          ),
                        },
                        a.name + ' 已恢复',
                        true,
                      )
                    }
                  >
                    <RotateCcw size={14} />
                    恢复
                  </Button>
                ) : (
                  <>
                    <Button
                      variant="ghost"
                      onClick={() => {
                        setDraft(structuredClone(a));
                        setTab('业务定义');
                        setError('');
                        setResult('');
                      }}
                    >
                      编辑
                    </Button>
                    <Button
                      variant="ghost"
                      onClick={() => {
                        setDraft(structuredClone(a));
                        setTab('测试');
                        setResult('');
                      }}
                    >
                      <Play size={13} />
                      测试
                    </Button>
                    <Button
                      variant="ghost"
                      size="icon"
                      aria-label={'复制 ' + a.name}
                      onClick={() =>
                        update(
                          {
                            agents: [
                              ...state.agents,
                              {
                                ...structuredClone(a),
                                id: crypto.randomUUID(),
                                name: a.name + ' 副本',
                                status: '草稿',
                                version: '0.1.0',
                                history: [],
                              },
                            ],
                          },
                          a.name + ' 已复制为独立草稿',
                          true,
                        )
                      }
                    >
                      <Copy size={14} />
                    </Button>
                    <Button
                      variant="ghost"
                      size="icon"
                      aria-label={'导出 ' + a.name}
                      onClick={() => download(a.id + '.json', a)}
                    >
                      <Download size={14} />
                    </Button>
                    <Button
                      variant="ghost"
                      size="icon"
                      aria-label={'软删除 ' + a.name}
                      onClick={() => setDeleteId(a.id)}
                    >
                      <Trash2 size={14} />
                    </Button>
                  </>
                )}
              </div>
            </article>
          ))}
        </div>
      )}
      <Modal
        title={draft?.name || '创建业务 Agent'}
        open={!!draft}
        onClose={() => setDraft(null)}
        description="业务定义与平台 Kernel 分离。修改会使当前评测与发布冻结失效。"
      >
        {draft && (
          <>
            <Tabs
              items={['业务定义', '能力与依赖', '版本', '测试']}
              value={tab}
              onChange={setTab}
            />
            {tab === '业务定义' && (
              <>
                <div className="two-col">
                  <Field label="Agent 名称">
                    <Input
                      maxLength={60}
                      value={draft.name}
                      onChange={(e) =>
                        setDraft({ ...draft, name: e.target.value })
                      }
                    />
                  </Field>
                  <Field label="业务 Role">
                    <Input
                      value={draft.role}
                      onChange={(e) =>
                        setDraft({ ...draft, role: e.target.value })
                      }
                    />
                  </Field>
                </div>
                <Field label="业务描述 / Context">
                  <Textarea
                    value={draft.description}
                    onChange={(e) =>
                      setDraft({ ...draft, description: e.target.value })
                    }
                  />
                </Field>
                <Field label="Knowledge">
                  <Input
                    value={draft.knowledge}
                    onChange={(e) =>
                      setDraft({ ...draft, knowledge: e.target.value })
                    }
                  />
                </Field>
                <Field label="Rule / Interaction Policy">
                  <Textarea
                    value={draft.interaction}
                    onChange={(e) =>
                      setDraft({ ...draft, interaction: e.target.value })
                    }
                  />
                </Field>
                <Field label="Permission">
                  <Input
                    value={draft.permission}
                    onChange={(e) =>
                      setDraft({ ...draft, permission: e.target.value })
                    }
                  />
                </Field>
                <Notice>
                  不可覆盖的平台策略：租户隔离、敏感数据脱敏、不可伪造工具结果、高风险写操作确认。
                </Notice>
              </>
            )}
            {tab === '能力与依赖' && (
              <>
                <p className="muted">
                  选择业务能力。应用打包时锁定具体版本；Agent 不直接保存 API
                  密钥。
                </p>
                <div className="cap-checks">
                  {capabilities.map((c) => (
                    <label key={c.id}>
                      <Checkbox
                        checked={draft.capabilities.includes(c.id)}
                        onCheckedChange={(v) =>
                          setDraft({
                            ...draft,
                            capabilities: v
                              ? [...draft.capabilities, c.id]
                              : draft.capabilities.filter((x) => x !== c.id),
                          })
                        }
                      />
                      <span>
                        {c.name}
                        <small>
                          {c.id} @ {c.version}
                        </small>
                      </span>
                      <Badge>{c.impl}</Badge>
                    </label>
                  ))}
                </div>
                <Notice>
                  上游：业务模型 BM-001；下游：
                  {state.application.selected.includes(draft.id)
                    ? state.application.name
                    : '尚未被应用引用'}
                  。
                </Notice>
              </>
            )}
            {tab === '版本' && (
              <>
                <Rows
                  headers={['版本', '变更说明']}
                  rows={[
                    [draft.version, '当前编辑版本'],
                    ...draft.history.map((h) => [h.version, h.note]),
                  ]}
                />
                <Button
                  variant="outline"
                  onClick={() => {
                    const parts = draft.version.split('.').map(Number);
                    const version = `${parts[0] || 0}.${(parts[1] || 0) + 1}.0`;
                    setDraft({
                      ...draft,
                      version,
                      status: '草稿',
                      history: [
                        { version: draft.version, note: '已保存的历史定义' },
                        ...draft.history,
                      ],
                    });
                  }}
                >
                  创建下一 Minor 版本
                </Button>
                <Notice>
                  版本保存在草稿中，点击“保存 Agent”后生效。生产版本保持冻结。
                </Notice>
              </>
            )}
            {tab === '测试' && (
              <>
                <Field label="业务测试输入">
                  <Textarea
                    value={test}
                    onChange={(e) => setTest(e.target.value)}
                  />
                </Field>
                <Button
                  disabled={busy || !test.trim()}
                  onClick={() => {
                    setBusy(true);
                    setTimeout(() => {
                      setResult(
                        draft.capabilities.length
                          ? `模拟运行：${draft.role}\n输入：${test}\n已加载 ${draft.capabilities.join(', ')}\nPermission → tenant:current 已校验\nBehavior Policy → 执行写操作前请求确认\n结果：等待用户确认 / 必要时转人工。未调用真实工具。`
                          : '运行被阻断：未配置 Capability。请先选择至少一个业务能力。',
                      );
                      setBusy(false);
                    }, 700);
                  }}
                >
                  <Play size={14} />
                  {busy ? '运行中…' : '运行模拟测试'}
                </Button>
                {result && <pre className="code-block">{result}</pre>}
                <Notice>
                  Agent 单次模拟测试不等于应用 Release
                  Gate；请在测试与评测中运行完整回归。
                </Notice>
              </>
            )}
            {error && <Notice tone="warning">{error}</Notice>}
            <div className="modal-actions">
              <Button variant="outline" onClick={() => setDraft(null)}>
                取消
              </Button>
              <Button onClick={save} disabled={draft.deleted}>
                <Save size={14} />
                保存 Agent
              </Button>
            </div>
          </>
        )}
      </Modal>
      <Modal
        title="软删除业务 Agent"
        open={!!deleteId}
        onClose={() => setDeleteId('')}
      >
        <p>Agent 将移入回收站。历史发布包仍保留冻结快照。</p>
        {state.application.selected.includes(deleteId) ? (
          <Notice tone="warning">
            此 Agent 被当前应用引用，不能删除。请先在应用构建器中解除引用。
          </Notice>
        ) : (
          <Button
            variant="destructive"
            onClick={() => {
              update(
                {
                  agents: state.agents.map((a) =>
                    a.id === deleteId ? { ...a, deleted: true } : a,
                  ),
                },
                'Agent 已移入回收站',
                true,
              );
              setDeleteId('');
            }}
          >
            确认软删除
          </Button>
        )}
      </Modal>
    </>
  );
}
