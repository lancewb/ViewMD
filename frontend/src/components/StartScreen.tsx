import { useRef } from 'react';
import { FileText, FolderOpen, History } from 'lucide-react';
import gsap from 'gsap';
import { useGSAP } from '@gsap/react';
import type { ProjectSummary } from '../types';

interface StartScreenProps {
  lastProject?: ProjectSummary | null;
  busy: boolean;
  onRestore: () => void;
  onChooseFolder: () => void;
}

export function StartScreen({ lastProject, busy, onRestore, onChooseFolder }: StartScreenProps) {
  const scope = useRef<HTMLDivElement>(null);

  useGSAP(() => {
    const mm = gsap.matchMedia();

    mm.add('(prefers-reduced-motion: no-preference)', () => {
      gsap.from('.start-animate', {
        autoAlpha: 0,
        y: 22,
        duration: 0.72,
        stagger: 0.08,
        ease: 'power3.out',
      });
    });

    return () => mm.revert();
  }, { scope });

  return (
    <main className="start-screen" ref={scope}>
      <section className="start-copy start-animate">
        <div className="brand-mark" aria-hidden="true">
          <FileText size={30} />
        </div>
        <p className="eyebrow">ViewMD</p>
        <h1>Markdown workspace</h1>
        <p className="start-lede">
          选择一个入口开始。
        </p>
      </section>

      <section className="start-actions start-animate" aria-label="Project actions">
        <button
          className="start-action primary"
          type="button"
          onClick={onRestore}
          disabled={!lastProject || busy}
        >
          <History size={22} />
          <span>
            <strong>恢复上次工程</strong>
            <small>{lastProject ? lastProject.path : '暂无可恢复的工程'}</small>
          </span>
        </button>

        <button className="start-action" type="button" onClick={onChooseFolder} disabled={busy}>
          <FolderOpen size={22} />
          <span>
            <strong>选择文件夹</strong>
            <small>打开现有目录或从空文件夹开始</small>
          </span>
        </button>
      </section>

    </main>
  );
}
