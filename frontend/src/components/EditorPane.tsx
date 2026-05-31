import { markdown } from '@codemirror/lang-markdown';
import { oneDark } from '@codemirror/theme-one-dark';
import { EditorView } from '@codemirror/view';
import CodeMirror from '@uiw/react-codemirror';

const editorTheme = EditorView.theme({
  '&': {
    backgroundColor: '#0b1114',
    color: '#f5f3e8',
    height: '100%',
    fontSize: '14px',
  },
  '.cm-scroller': {
    fontFamily: '"Cascadia Code", "JetBrains Mono", "Fira Code", monospace',
    lineHeight: '1.68',
  },
  '.cm-content': {
    padding: '18px 0 28px',
    caretColor: '#f0b443',
  },
  '.cm-gutters': {
    backgroundColor: '#090e11',
    color: '#9fb2a6',
    borderRight: '1px solid rgba(255,255,255,0.1)',
  },
  '.cm-activeLineGutter': {
    backgroundColor: 'rgba(240, 180, 67, 0.16)',
  },
  '.cm-activeLine': {
    backgroundColor: 'rgba(240, 180, 67, 0.1)',
  },
  '.cm-cursor': {
    borderLeftColor: '#f0b443',
  },
  '.cm-selectionBackground, &.cm-focused .cm-selectionBackground': {
    backgroundColor: 'rgba(57, 194, 160, 0.42)',
  },
  '.cm-content ::selection': {
    backgroundColor: 'rgba(57, 194, 160, 0.62)',
    color: '#fffaf0',
  },
});

interface EditorPaneProps {
  fileName: string;
  value: string;
  onChange: (value: string) => void;
}

export function EditorPane({ fileName, value, onChange }: EditorPaneProps) {
  return (
    <section className="editor-pane" aria-label="Markdown editor">
      <div className="pane-titlebar">
        <span>{fileName}</span>
      </div>
      <CodeMirror
        className="code-editor"
        value={value}
        height="100%"
        theme={oneDark}
        basicSetup={{
          autocompletion: true,
          bracketMatching: true,
          closeBrackets: true,
          foldGutter: true,
          highlightActiveLine: true,
          highlightActiveLineGutter: true,
          lineNumbers: true,
        }}
        extensions={[markdown(), EditorView.lineWrapping, editorTheme]}
        onChange={onChange}
      />
    </section>
  );
}
