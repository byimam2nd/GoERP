import React, { useEffect, useState } from 'react';

interface DocField {
  name: string;
  label: string;
  fieldtype: string;
  in_list_view?: boolean;
}

interface DocType {
  name: string;
  fields: DocField[];
}

interface DynamicListProps {
  doctypeName: string;
  tenantID: string;
  token: string;
}

const DynamicList: React.FC<DynamicListProps> = ({ doctypeName, tenantID, token }) => {
  const [meta, setMeta] = useState<DocType | null>(null);
  const [data, setData] = useState<any[]>([]);
  const [loading, setLoading] = useState(true);

  useEffect(() => {
    // 1. Fetch Metadata
    const fetchMeta = fetch(`/api/v1/meta/${doctypeName}`, {
      headers: { 'X-GoERP-Tenant': tenantID, 'Authorization': `Bearer ${token}` }
    }).then(res => res.json());

    // 2. Fetch Data
    const fetchData = fetch(`/api/v1/resource/${doctypeName}`, {
      headers: { 'X-GoERP-Tenant': tenantID, 'Authorization': `Bearer ${token}` }
    }).then(res => res.json());

    Promise.all([fetchMeta, fetchData]).then(([metaData, listData]) => {
      setMeta(metaData);
      setData(listData.data || []);
      setLoading(false);
    });
  }, [doctypeName]);

  if (loading || !meta) return <div className="p-8 text-center animate-pulse">Scanning Registry...</div>;

  // Filter fields that should appear in list view
  const listFields = meta.fields.filter(f => f.in_list_view && !['Section Break', 'Column Break'].includes(f.fieldtype));
  
  // Fallback if no fields marked as in_list_view, show first 3 Data fields
  const displayFields = listFields.length > 0 ? listFields : 
    meta.fields.filter(f => f.fieldtype === 'Data').slice(0, 3);

  return (
    <div className="max-w-7xl mx-auto p-6">
      <div className="flex justify-between items-center mb-6">
        <h1 className="text-3xl font-black text-gray-900 tracking-tight">{meta.name} List</h1>
        <button className="bg-blue-600 text-white px-5 py-2 rounded-lg font-bold shadow-md hover:bg-blue-700 transition">
          + New {meta.name}
        </button>
      </div>

      <div className="bg-white border rounded-xl shadow-sm overflow-hidden">
        <table className="w-full text-left border-collapse">
          <thead className="bg-gray-50 border-b">
            <tr>
              <th className="px-6 py-4 text-sm font-bold text-gray-600 uppercase">ID</th>
              {displayFields.map(f => (
                <th key={f.name} className="px-6 py-4 text-sm font-bold text-gray-600 uppercase">
                  {f.label}
                </th>
              ))}
              <th className="px-6 py-4 text-sm font-bold text-gray-600 uppercase text-right">Modified</th>
            </tr>
          </thead>
          <tbody className="divide-y">
            {data.map((row, idx) => (
              <tr key={idx} className="hover:bg-blue-50 transition cursor-pointer group">
                <td className="px-6 py-4 font-mono text-xs text-blue-600 font-bold">
                  {row.name}
                </td>
                {displayFields.map(f => (
                  <td key={f.name} className="px-6 py-4 text-sm text-gray-700 font-medium">
                    {row[f.name]?.toString() || '-'}
                  </td>
                ))}
                <td className="px-6 py-4 text-xs text-gray-400 text-right">
                   {new Date(row.modified).toLocaleDateString()}
                </td>
              </tr>
            ))}
          </tbody>
        </table>
        
        {data.length === 0 && (
          <div className="p-20 text-center text-gray-400 font-medium italic">
            No data found for this DocType.
          </div>
        )}
      </div>
    </div>
  );
};

export default DynamicList;
