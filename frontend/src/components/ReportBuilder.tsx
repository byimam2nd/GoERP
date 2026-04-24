import React, { useEffect, useState } from 'react';

interface ReportBuilderProps {
  doctypeName: string;
  tenantID: string;
  token: string;
}

const ReportBuilder: React.FC<ReportBuilderProps> = ({ doctypeName, tenantID, token }) => {
  const [fields, setFields] = useState<any[]>([]);
  const [selectedFields, setSelectedFields] = useState<string[]>(['name']);
  const [reportData, setReportData] = useState<any[]>([]);
  const [loading, setLoading] = useState(false);

  useEffect(() => {
    // Fetch DocType Meta to get available fields
    fetch(`/api/v1/meta/${doctypeName}`, {
      headers: { 'X-GoERP-Tenant': tenantID, 'Authorization': `Bearer ${token}` }
    })
    .then(res => res.json())
    .then(data => setFields(data.fields.filter((f: any) => !['Section Break', 'Column Break'].includes(f.fieldtype))));
  }, [doctypeName]);

  const runReport = async () => {
    setLoading(true);
    const res = await fetch('/api/v1/report', {
      method: 'POST',
      headers: {
        'Content-Type': 'application/json',
        'X-GoERP-Tenant': tenantID,
        'Authorization': `Bearer ${token}`
      },
      body: JSON.stringify({
        doctype: doctypeName,
        fields: selectedFields,
        filters: {},
        limit: 50
      })
    });
    const result = await res.json();
    setReportData(result.data || []);
    setLoading(false);
  };

  return (
    <div className="flex h-screen bg-gray-100">
      {/* Sidebar: Config */}
      <aside className="w-80 bg-white border-r p-6 overflow-y-auto">
        <h2 className="text-xl font-black mb-6 flex items-center space-x-2">
           <span className="bg-blue-600 w-2 h-8 rounded"></span>
           <span>Report Config</span>
        </h2>
        
        <div className="space-y-6">
          <div>
            <label className="text-xs font-bold text-gray-400 uppercase tracking-widest block mb-4">Select Columns</label>
            <div className="space-y-2">
              {fields.map(f => (
                <label key={f.name} className="flex items-center space-x-3 p-2 hover:bg-blue-50 rounded-lg cursor-pointer transition">
                  <input 
                    type="checkbox" 
                    checked={selectedFields.includes(f.name)}
                    className="w-5 h-5 rounded text-blue-600 border-gray-300"
                    onChange={(e) => {
                       if (e.target.checked) setSelectedFields([...selectedFields, f.name]);
                       else setSelectedFields(selectedFields.filter(sf => sf !== f.name));
                    }}
                  />
                  <span className="text-sm font-medium text-gray-700">{f.label}</span>
                </label>
              ))}
            </div>
          </div>

          <button 
            onClick={runReport}
            className="w-full bg-blue-600 text-white py-3 rounded-xl font-bold shadow-lg hover:bg-blue-700 transition transform active:scale-95 flex items-center justify-center space-x-2"
          >
            <span>🚀 Run Report</span>
          </button>
        </div>
      </aside>

      {/* Main: Table */}
      <main className="flex-1 p-8 overflow-auto">
        <div className="bg-white rounded-2xl shadow-xl border border-gray-100 overflow-hidden">
          <table className="w-full text-left border-collapse">
            <thead className="bg-gray-50 border-b">
              <tr>
                {selectedFields.map(f => (
                  <th key={f} className="px-6 py-4 text-xs font-black text-gray-500 uppercase tracking-widest">{f}</th>
                ))}
              </tr>
            </thead>
            <tbody className="divide-y">
              {reportData.map((row, idx) => (
                <tr key={idx} className="hover:bg-gray-50 transition">
                  {selectedFields.map(f => (
                    <td key={f} className="px-6 py-4 text-sm text-gray-700 font-medium">
                      {row[f]?.toString() || '-'}
                    </td>
                  ))}
                </tr>
              ))}
            </tbody>
          </table>
          
          {reportData.length === 0 && !loading && (
            <div className="p-20 text-center text-gray-400 italic">
               Select columns and click "Run Report" to generate analytics.
            </div>
          )}
          {loading && (
            <div className="p-20 text-center text-blue-600 animate-pulse font-bold">
               Generating Report...
            </div>
          )}
        </div>
      </main>
    </div>
  );
};

export default ReportBuilder;
