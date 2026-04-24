import React, { useState } from 'react';

const DocTypeBuilder: React.FC<{ tenantID: string, token: string }> = ({ tenantID, token }) => {
  const [docMeta, setDocMeta] = useState({ name: '', module: 'custom', is_submittable: false });
  const [fields, setFields] = useState<any[]>([]);

  const addField = () => {
    setFields([...fields, { fieldname: '', label: '', fieldtype: 'Data', in_list_view: true }]);
  };

  const saveDocType = async () => {
    // 1. Create the DocType Record
    const res = await fetch('/api/v1/resource/DocType', {
      method: 'POST',
      headers: { 'Content-Type': 'application/json', 'X-GoERP-Tenant': tenantID, 'Authorization': `Bearer ${token}` },
      body: JSON.stringify(docMeta)
    });

    if (!res.ok) return alert('Failed to create DocType header');

    // 2. Create the Fields
    for (const field of fields) {
      await fetch('/api/v1/resource/DocField', {
        method: 'POST',
        headers: { 'Content-Type': 'application/json', 'X-GoERP-Tenant': tenantID, 'Authorization': `Bearer ${token}` },
        body: JSON.stringify({ ...field, parent: docMeta.name })
      });
    }

    alert(`Success! DocType "${docMeta.name}" is now LIVE.`);
  };

  return (
    <div className="max-w-5xl mx-auto p-10 bg-white shadow-2xl rounded-3xl border border-gray-100">
      <h1 className="text-4xl font-black text-gray-900 mb-8 tracking-tighter italic">DocType Architect</h1>
      
      <div className="grid grid-cols-2 gap-8 mb-10 p-6 bg-gray-50 rounded-2xl border border-gray-200">
        <input 
          placeholder="New DocType Name (e.g. Asset)" 
          className="p-4 rounded-xl border-2 border-gray-200 focus:border-blue-600 outline-none font-bold"
          onChange={e => setDocMeta({...docMeta, name: e.target.value})}
        />
        <input 
          placeholder="Module Name" 
          className="p-4 rounded-xl border-2 border-gray-200 focus:border-blue-600 outline-none"
          value={docMeta.module}
          onChange={e => setDocMeta({...docMeta, module: e.target.value})}
        />
      </div>

      <div className="space-y-4">
        <div className="flex justify-between items-center">
           <h2 className="text-xl font-bold text-gray-700">Fields Definition</h2>
           <button onClick={addField} className="text-blue-600 font-bold hover:underline">+ Add Row</button>
        </div>

        {fields.map((f, idx) => (
          <div key={idx} className="flex space-x-4 items-end bg-white p-4 border rounded-xl shadow-sm">
             <div className="flex-1">
               <label className="text-[10px] font-black text-gray-400 uppercase">Fieldname</label>
               <input 
                 className="w-full border-b-2 outline-none p-1 focus:border-blue-500"
                 onChange={e => {
                   const newFields = [...fields];
                   newFields[idx].fieldname = e.target.value;
                   setFields(newFields);
                 }}
               />
             </div>
             <div className="flex-1">
               <label className="text-[10px] font-black text-gray-400 uppercase">Label</label>
               <input 
                 className="w-full border-b-2 outline-none p-1 focus:border-blue-500"
                 onChange={e => {
                   const newFields = [...fields];
                   newFields[idx].label = e.target.value;
                   setFields(newFields);
                 }}
               />
             </div>
             <div className="w-40">
               <label className="text-[10px] font-black text-gray-400 uppercase">Type</label>
               <select 
                 className="w-full border-b-2 outline-none p-1"
                 onChange={e => {
                   const newFields = [...fields];
                   newFields[idx].fieldtype = e.target.value;
                   setFields(newFields);
                 }}
               >
                 <option>Data</option><option>Int</option><option>Check</option><option>Text</option><option>Select</option>
               </select>
             </div>
          </div>
        ))}
      </div>

      <button 
        onClick={saveDocType}
        className="mt-12 w-full bg-black text-white p-5 rounded-2xl font-black text-xl shadow-2xl hover:bg-gray-800 transition"
      >
        🚀 Build and Migrate DocType
      </button>
    </div>
  );
};

export default DocTypeBuilder;
