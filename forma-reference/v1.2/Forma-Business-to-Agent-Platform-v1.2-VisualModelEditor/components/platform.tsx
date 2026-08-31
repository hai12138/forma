'use client';
import { useState, useEffect } from 'react';
import {
  Search,
  ChevronDown,
  Bell,
  Layers,
  Check,
  ArrowRight,
} from 'lucide-react';
import { Input } from '@/components/ui/input';
import { navigation } from '@/lib/navigation';
import { StoreProvider, useStore } from '@/lib/store';
import { Modal } from './shared';
import Overview from './screens/overview';
import { Analyst, Business } from './screens/business';
import DataPlane from './screens/data';
import Capabilities from './screens/capabilities';
import Agents from './screens/agents';
import Applications from './screens/applications';
import { Evaluation, Releases, HumanTasks } from './screens/ship';
import {
  Channels,
  Runtime,
  Observability,
  Governance,
} from './screens/operate';
import { Delivery, DesignSystem } from './screens/delivery-design';
export default function Platform() {
  return (
    <StoreProvider>
      <Shell />
    </StoreProvider>
  );
}
function Shell() {
  const { state, toast, notify } = useStore();
  const [route, setRoute] = useState('overview');
  const [search, setSearch] = useState(false);
  const [query, setQuery] = useState('');
  const items = navigation.flatMap((g) => g.items);
  const item = items.find((i) => i.id === route);
  useEffect(() => {
    const read = () =>
      setRoute(window.location.pathname.split('/')[1] || 'overview');
    read();
    window.addEventListener('popstate', read);
    const key = (e: KeyboardEvent) => {
      if ((e.metaKey || e.ctrlKey) && e.key === 'k') {
        e.preventDefault();
        setSearch((v) => !v);
      }
    };
    window.addEventListener('keydown', key);
    return () => {
      window.removeEventListener('popstate', read);
      window.removeEventListener('keydown', key);
    };
  }, []);
  const go = (r: string) => {
    setRoute(r);
    window.history.pushState({}, '', r === 'overview' ? '/' : '/' + r);
    window.scrollTo({ top: 0 });
    setSearch(false);
    document.title =
      (items.find((i) => i.id === r)?.label || '页面未找到') + ' · Forma';
  };
  const screen = () => {
    switch (route) {
      case 'overview':
        return <Overview go={go} />;
      case 'analyst':
        return <Analyst go={go} />;
      case 'business':
        return <Business go={go} />;
      case 'data':
        return <DataPlane />;
      case 'capabilities':
        return <Capabilities />;
      case 'agents':
        return <Agents />;
      case 'applications':
        return <Applications go={go} />;
      case 'human':
        return <HumanTasks />;
      case 'evaluation':
        return <Evaluation go={go} />;
      case 'releases':
        return <Releases go={go} />;
      case 'channels':
        return <Channels />;
      case 'runtime':
        return <Runtime />;
      case 'observability':
        return <Observability />;
      case 'governance':
        return <Governance />;
      case 'delivery':
        return <Delivery go={go} />;
      case 'design':
        return <DesignSystem />;
      default:
        return (
          <div className="empty">
            <h1>页面未找到</h1>
            <button onClick={() => go('overview')}>返回总览</button>
          </div>
        );
    }
  };
  return (
    <div className="platform">
      <a className="skip-link" href="#main-content">
        跳转至主要内容
      </a>
      <aside className="sidebar">
        <button
          className="brand"
          onClick={() => go('overview')}
          aria-label="Forma 总览"
        >
          <span className="brand-mark">
            <Layers size={22} />
          </span>
          forma<span className="brand-dot">®</span>
        </button>
        <button
          className="workspace"
          onClick={() =>
            notify('当前为 Northstar 单租户演示空间；生产租户切换需重新鉴权。')
          }
        >
          <span className="workspace-icon">N</span>
          <span>
            Northstar 园区<small>Enterprise workspace</small>
          </span>
          <ChevronDown size={14} />
        </button>
        <nav aria-label="主要导航">
          {navigation.map((g) => (
            <div className="nav-group" key={g.group}>
              <p>{g.group}</p>
              {g.items.map((i) => (
                <a
                  href={i.id === 'overview' ? '/' : '/' + i.id}
                  title={i.label}
                  key={i.id}
                  onClick={(e) => {
                    if (!e.metaKey && !e.ctrlKey) {
                      e.preventDefault();
                      go(i.id);
                    }
                  }}
                  aria-current={route === i.id ? 'page' : undefined}
                  className={route === i.id ? 'nav-item active' : 'nav-item'}
                >
                  <i.icon size={17} />
                  <span className="nav-label">{i.label}</span>
                  {i.id === 'human' && (
                    <span className="nav-count">
                      {
                        Object.values(state.human).filter((s) => s === '待处理')
                          .length
                      }
                    </span>
                  )}
                </a>
              ))}
            </div>
          ))}
        </nav>
        <div className="sidebar-bottom">
          <span className="avatar">董</span>
          <span>
            董志海<small>演示用户 · 管理员</small>
          </span>
          <ChevronDown size={14} />
        </div>
      </aside>
      <div className="main-shell">
        <header className="topbar">
          <div className="breadcrumb">
            工作空间 <span>/</span> <strong>{item?.label || '未找到'}</strong>
          </div>
          <div className="top-actions">
            <span className="environment">
              <i />
              Prototype · {state.stage}
            </span>
            <button
              aria-label="全局搜索 Ctrl K"
              onClick={() => setSearch(true)}
            >
              <Search size={18} />
              <kbd>⌘ K</kbd>
            </button>
            <button aria-label="查看待办通知" onClick={() => go('human')}>
              <Bell size={18} />
            </button>
            <span className="avatar small">董</span>
          </div>
        </header>
        <main id="main-content" key={route}>
          {screen()}
          <footer>
            FORMA / BUSINESS ENGINEERING{' '}
            <span>
              交互原型 · 外部调用均为模拟 · 当前草稿 r{state.revision}
            </span>
          </footer>
        </main>
      </div>
      {toast && (
        <output className="toast">
          <Check size={17} />
          {toast}
        </output>
      )}
      <Modal
        title="搜索工作空间"
        description="快速前往资产、治理与交付页面。支持 Ctrl / ⌘ K。"
        open={search}
        onClose={() => setSearch(false)}
      >
        <Input
          aria-label="搜索页面"
          value={query}
          placeholder="输入业务、Agent、发布、数据…"
          onChange={(e) => setQuery(e.target.value)}
        />
        <div className="command-results">
          {items
            .filter((i) =>
              (i.label + i.id).toLowerCase().includes(query.toLowerCase()),
            )
            .map((i) => (
              <button key={i.id} onClick={() => go(i.id)}>
                <i.icon size={17} />
                {i.label}
                <ArrowRight size={14} />
              </button>
            ))}
          {!items.some((i) =>
            (i.label + i.id).toLowerCase().includes(query.toLowerCase()),
          ) && <p className="muted">未找到匹配页面。</p>}
        </div>
      </Modal>
    </div>
  );
}
