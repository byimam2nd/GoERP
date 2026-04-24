import React, { useEffect, useState } from 'react';

interface Link {
  parent_doctype: string;
  parent_name: string;
  child_doctype: string;
  child_name: string;
}

const DocConnections: React.FC<{ doctype: string, name: string, tenantID: string, token: string }> = ({ doctype, name, tenantID, token }) => {
  const [links, setLinks] = useState<{ parents: Link[], children: Link[] }>({ parents: [], children: [] });

  useEffect(() => {
    fetch(`/api/v1/resource/${doctype}/${name}/links`, {
      headers: { 'X-GoERP-Tenant': tenantID, 'Authorization': `Bearer ${token}` }
    })
    .then(res => res.json())
    .then(data => setLinks(data));
  }, [doctype, name]);

  if (links.parents.length === 0 && links.children.length === 0) return null;

  return (
    <div className="mb-6 p-4 bg-gray-50 rounded-2xl border border-gray-100 flex flex-wrap gap-4 items-center">
      <span className="text-[10px] font-black text-gray-400 uppercase tracking-widest">Connections</span>
      
      {/* Upward Connections */}
      {links.parents.map((l, i) => (
        <div key={i} className="flex items-center space-x-2 bg-white px-3 py-1.5 rounded-full border shadow-sm hover:border-blue-500 cursor-pointer transition group">
           <span className="text-xs text-gray-400">←</span>
           <span className="text-xs font-bold text-gray-600 group-hover:text-blue-600">{l.parent_doctype}: {l.parent_name}</span>
        </div>
      ))}

      {/* Downward Connections */}
      {links.children.map((l, i) => (
        <div key={i} className="flex items-center space-x-2 bg-white px-3 py-1.5 rounded-full border shadow-sm hover:border-green-500 cursor-pointer transition group">
           <span className="text-xs text-gray-600 group-hover:text-green-600 font-bold">{l.child_doctype}: {l.child_name}</span>
           <span className="text-xs text-gray-400">→</span>
        </div>
      ))}
    </div>
  );
};

export default DocConnections;
