import React, { useState, useEffect, useCallback } from 'react';

const GlobalSearch: React.FC<any> = ({ tenantID, token, onNavigate }) => {
  const [isOpen, setIsOpen] = useState(false);
  const [query, setQuery] = useState('');
  const [results, setResults] = useState<any[]>([]);
  const [loading, setLoading] = useState(false);
  const [selectedIndex, setSelectedIndex] = useState(0);

  const toggle = useCallback(() => setIsOpen(prev => !prev), []);

  // Shortcut Listener (Ctrl+K or Cmd+K)
  useEffect(() => {
    const handleKeyDown = (e: KeyboardEvent) => {
      if ((e.ctrlKey || e.metaKey) && e.key === 'k') {
        e.preventDefault();
        toggle();
      }
      if (e.key === 'Escape') setIsOpen(false);
    };
    window.addEventListener('keydown', handleKeyDown);
    return () => window.removeEventListener('keydown', handleKeyDown);
  }, [toggle]);

  useEffect(() => {
    if (!isOpen || query.length < 2) {
      setResults([]);
      return;
    }

    const t = setTimeout(() => {
      setLoading(true);
      fetch(`/api/v1/search?q=${query}`, {
        headers: { 'X-GoERP-Tenant': tenantID, 'Authorization': `Bearer ${token}` }
      })
        .then(res => res.json())
        .then(data => {
          setResults(data || []);
          setLoading(false);
          setSelectedIndex(0);
        });
    }, 200);

    return () => clearTimeout(t);
  }, [query, isOpen, tenantID, token]);

  const handleSelect = (result: any) => {
    setIsOpen(false);
    setQuery('');
    // Trigger navigation callback
    if (onNavigate) onNavigate(result.route, result);
  };

  if (!isOpen) return null;

  return (
    <div className="fixed inset-0 z-[100] flex items-start justify-center pt-20 bg-gray-900/50 backdrop-blur-sm animate-in fade-in duration-200">
      <div className="w-full max-w-2xl bg-white rounded-3xl shadow-2xl overflow-hidden border border-indigo-100 flex flex-col animate-in slide-in-from-top-4 duration-300">
        
        {/* Input Bar */}
        <div className="relative flex items-center p-6 border-b border-gray-100">
           <svg className="w-6 h-6 text-indigo-500 mr-4" fill="none" stroke="currentColor" viewBox="0 0 24 24"><path strokeLinecap="round" strokeLinejoin="round" strokeWidth={3} d="M21 21l-6-6m2-5a7 7 0 11-14 0 7 7 0 0114 0z" /></svg>
           <input 
             autoFocus
             placeholder="Search for anything... (e.g. INV-2024, iPhone, Customer Name)"
             className="w-full text-xl font-bold text-gray-800 outline-none placeholder-gray-300"
             value={query}
             onChange={e => setQuery(e.target.value)}
             onKeyDown={e => {
               if (e.key === 'ArrowDown') setSelectedIndex(p => (p + 1) % results.length);
               if (e.key === 'ArrowUp') setSelectedIndex(p => (p - 1 + results.length) % results.length);
               if (e.key === 'Enter') handleSelect(results[selectedIndex]);
             }}
           />
           <div className="text-[10px] font-black text-gray-300 uppercase tracking-widest bg-gray-50 px-2 py-1 rounded-lg">ESC to close</div>
        </div>

        {/* Results Area */}
        <div className="max-h-[450px] overflow-y-auto bg-gray-50/50 p-2">
          {loading && <div className="p-10 text-center animate-pulse text-indigo-600 font-black tracking-widest text-xs uppercase">Analyzing System...</div>}
          {!loading && results.length === 0 && query.length >= 2 && <div className="p-10 text-center text-gray-400 font-medium">No records matching "{query}" found.</div>}
          {!loading && query.length < 2 && <div className="p-10 text-center text-gray-300 font-bold text-xs uppercase tracking-[0.2em]">Start typing to explore GoERP...</div>}
          
          <div className="space-y-1">
            {results.map((r, idx) => (
              <div 
                key={idx}
                onClick={() => handleSelect(r)}
                onMouseEnter={() => setSelectedIndex(idx)}
                className={`p-4 rounded-2xl cursor-pointer transition flex items-center justify-between ${idx === selectedIndex ? 'bg-indigo-600 text-white shadow-xl shadow-indigo-100' : 'hover:bg-white hover:shadow-md'}`}
              >
                <div className="flex items-center">
                   <div className={`w-10 h-10 rounded-xl flex items-center justify-center mr-4 font-black text-xs ${idx === selectedIndex ? 'bg-indigo-500' : 'bg-white shadow-sm text-indigo-600 border border-indigo-50'}`}>
                      {r.doctype.substring(0, 2).toUpperCase()}
                   </div>
                   <div>
                     <div className={`font-black tracking-tight ${idx === selectedIndex ? 'text-white' : 'text-gray-900'}`}>{r.name}</div>
                     <div className={`text-[10px] font-bold uppercase tracking-wider ${idx === selectedIndex ? 'text-indigo-200' : 'text-gray-400'}`}>{r.description}</div>
                   </div>
                </div>
                <div className={`text-[10px] font-black uppercase tracking-widest px-3 py-1 rounded-full ${idx === selectedIndex ? 'bg-indigo-500 text-white' : 'bg-indigo-50 text-indigo-400'}`}>
                  {r.doctype}
                </div>
              </div>
            ))}
          </div>
        </div>

        {/* Footer */}
        <div className="p-4 bg-white border-t border-gray-100 flex items-center justify-between">
           <div className="flex space-x-6">
              <div className="flex items-center text-[10px] font-black text-gray-300 uppercase tracking-widest">
                 <span className="bg-gray-100 px-1.5 py-0.5 rounded mr-2 border">↵</span> Select
              </div>
              <div className="flex items-center text-[10px] font-black text-gray-300 uppercase tracking-widest">
                 <span className="bg-gray-100 px-1.5 py-0.5 rounded mr-2 border">↓↑</span> Navigate
              </div>
           </div>
           <div className="text-[10px] font-black text-indigo-300 uppercase italic tracking-tighter">Powered by GoERP Search Engine</div>
        </div>
      </div>
    </div>
  );
};

export default GlobalSearch;
