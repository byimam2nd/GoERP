import React, { useState } from 'react';

interface CustomFieldBuilderProps {
  tenantID: string;
  token: string;
}

const CustomFieldBuilder: React.FC<CustomFieldBuilderProps> = ({ tenantID, token }) => {
  const [field, setField] = useState({
    dt: '',
    label: '',
    fieldname: '',
    field_type: 'Data',
    insert_after: '',
    in_list_view: false,
    is_required: false
  });

  const fieldTypes = ["Data", "Int", "Float", "Currency", "Date", "DateTime", "Check", "Link", "Select", "Text", "Section Break", "Column Break"];

  const handleSubmit = async (e: React.FormEvent) => {
    e.preventDefault();
    const res = await fetch('/api/v1/resource/CustomField', {
      method: 'POST',
      headers: {
        'Content-Type': 'application/json',
        'X-GoERP-Tenant': tenantID,
        'Authorization': `Bearer ${token}`
      },
      body: JSON.stringify(field)
    });

    if (res.ok) {
      alert('Magic! Field created and Database migrated.');
    } else {
      const err = await res.json();
      alert('Error: ' + err.error);
    }
  };

  return (
    <div className="max-w-4xl mx-auto p-8 bg-gradient-to-br from-indigo-50 to-white shadow-2xl rounded-3xl border border-indigo-100">
      <div className="flex items-center space-x-4 mb-8">
        <div className="bg-indigo-600 p-3 rounded-2xl shadow-lg">
          <svg className="w-8 h-8 text-white" fill="none" stroke="currentColor" viewBox="0 0 24 24">
            <path strokeLinecap="round" strokeLinejoin="round" strokeWidth="2" d="M12 6V4m0 2a2 2 0 100 4m0-4a2 2 0 110 4m-6 8a2 2 0 100-4m0 4a2 2 0 110-4m0 4v2m0-6V4m6 6v10m6-2a2 2 0 100-4m0 4a2 2 0 110-4m0 4v2m0-6V4" />
          </svg>
        </div>
        <div>
          <h1 className="text-3xl font-black text-gray-900 tracking-tight">Custom Field Builder</h1>
          <p className="text-indigo-600 font-medium italic text-sm">Design your database without coding</p>
        </div>
      </div>

      <form onSubmit={handleSubmit} className="grid grid-cols-2 gap-6">
        <div className="flex flex-col">
          <label className="text-xs font-bold text-gray-400 uppercase tracking-widest mb-2">Target DocType</label>
          <input 
            placeholder="e.g. Employee"
            className="border-2 border-gray-100 p-3 rounded-xl focus:border-indigo-500 outline-none transition"
            onChange={e => setField({...field, dt: e.target.value})}
          />
        </div>

        <div className="flex flex-col">
          <label className="text-xs font-bold text-gray-400 uppercase tracking-widest mb-2">Field Label</label>
          <input 
            placeholder="e.g. Mobile Number"
            className="border-2 border-gray-100 p-3 rounded-xl focus:border-indigo-500 outline-none transition"
            onChange={e => setField({...field, label: e.target.value, fieldname: e.target.value.toLowerCase().replace(/ /g, '_')})}
          />
        </div>

        <div className="flex flex-col">
          <label className="text-xs font-bold text-gray-400 uppercase tracking-widest mb-2">Field Type</label>
          <select 
            className="border-2 border-gray-100 p-3 rounded-xl focus:border-indigo-500 outline-none transition"
            onChange={e => setField({...field, field_type: e.target.value})}
          >
            {fieldTypes.map(t => <option key={t} value={t}>{t}</option>)}
          </select>
        </div>

        <div className="flex flex-col">
          <label className="text-xs font-bold text-gray-400 uppercase tracking-widest mb-2">Insert After</label>
          <input 
            placeholder="fieldname (optional)"
            className="border-2 border-gray-100 p-3 rounded-xl focus:border-indigo-500 outline-none transition"
            onChange={e => setField({...field, insert_after: e.target.value})}
          />
        </div>

        <div className="flex items-center space-x-6 p-4 bg-white rounded-2xl border-2 border-gray-50 col-span-2">
          <label className="flex items-center space-x-3 cursor-pointer">
            <input 
              type="checkbox" 
              className="w-6 h-6 rounded-lg text-indigo-600 border-gray-300 focus:ring-indigo-500" 
              onChange={e => setField({...field, in_list_view: e.target.checked})}
            />
            <span className="font-bold text-gray-700">Show in List View</span>
          </label>
          
          <label className="flex items-center space-x-3 cursor-pointer">
            <input 
              type="checkbox" 
              className="w-6 h-6 rounded-lg text-indigo-600 border-gray-300 focus:ring-indigo-500"
              onChange={e => setField({...field, is_required: e.target.checked})}
            />
            <span className="font-bold text-gray-700">Mandatory</span>
          </label>
        </div>

        <button 
          type="submit" 
          className="col-span-2 bg-indigo-600 text-white p-4 rounded-2xl font-black text-lg shadow-xl hover:bg-indigo-700 transform transition active:scale-95 flex justify-center items-center space-x-2"
        >
          <span>🔥 Create Magic Field</span>
        </button>
      </form>
    </div>
  );
};

export default CustomFieldBuilder;
