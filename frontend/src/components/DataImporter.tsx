import React, { useState } from 'react';

const DataImporter: React.FC<{ doctype: string, tenantID: string, token: string }> = ({ doctype, tenantID, token }) => {
  const [file, setFile] = useState<File | null>(null);
  const [loading, setLoading] = useState(false);
  const [result, setResult] = useState<any>(null);

  const downloadTemplate = () => {
    window.open(`/api/v1/import/template/${doctype}?token=${token}`, '_blank');
  };

  const handleUpload = async () => {
    if (!file) return;
    setLoading(true);
    
    const formData = new FormData();
    formData.append('file', file);

    const res = await fetch(`/api/v1/import/upload/${doctype}`, {
      method: 'POST',
      headers: {
        'X-GoERP-Tenant': tenantID,
        'Authorization': `Bearer ${token}`
      },
      body: formData
    });

    const data = await res.json();
    setResult(data);
    setLoading(false);
  };

  return (
    <div className="max-w-4xl mx-auto p-12 bg-white shadow-2xl rounded-[3rem] border border-gray-100 mt-10">
      <div className="text-center mb-12">
        <h1 className="text-5xl font-black text-gray-900 tracking-tighter mb-4">Data Importer</h1>
        <p className="text-gray-500 font-medium">Mass onboarding for <span className="text-blue-600 font-bold uppercase">{doctype}</span></p>
      </div>

      <div className="grid grid-cols-1 md:grid-cols-2 gap-8 mb-12">
         {/* Step 1 */}
         <div className="p-8 bg-blue-50 rounded-3xl border-2 border-blue-100 flex flex-col items-center text-center group hover:bg-blue-600 transition duration-500">
            <div className="w-16 h-16 bg-blue-600 text-white rounded-2xl flex items-center justify-center mb-6 shadow-lg group-hover:bg-white group-hover:text-blue-600 transition">
               <span className="text-2xl font-black">1</span>
            </div>
            <h3 className="text-xl font-bold mb-3 group-hover:text-white">Download Template</h3>
            <p className="text-sm text-blue-600 group-hover:text-blue-100 mb-6 font-medium">Get a CSV file with correct headers for this DocType.</p>
            <button 
              onClick={downloadTemplate}
              className="mt-auto px-8 py-3 bg-blue-600 text-white rounded-xl font-bold shadow-md hover:bg-blue-700 group-hover:bg-white group-hover:text-blue-600 transition"
            >
              Get CSV Template
            </button>
         </div>

         {/* Step 2 */}
         <div className="p-8 bg-gray-50 rounded-3xl border-2 border-gray-200 flex flex-col items-center text-center group hover:border-blue-500 transition duration-500">
            <div className="w-16 h-16 bg-gray-900 text-white rounded-2xl flex items-center justify-center mb-6 shadow-lg group-hover:bg-blue-600 transition">
               <span className="text-2xl font-black">2</span>
            </div>
            <h3 className="text-xl font-bold mb-3">Upload Data</h3>
            <p className="text-sm text-gray-500 mb-6 font-medium">Upload your filled CSV file to start the bulk insertion.</p>
            
            <input 
              type="file" 
              accept=".csv" 
              className="hidden" 
              id="file-upload"
              onChange={(e) => setFile(e.target.files?.[0] || null)}
            />
            <label 
              htmlFor="file-upload"
              className="mb-4 cursor-pointer text-sm font-bold text-gray-400 border-2 border-dashed border-gray-300 p-4 w-full rounded-2xl hover:border-blue-400 transition"
            >
              {file ? file.name : 'Click to select CSV'}
            </label>

            <button 
              onClick={handleUpload}
              disabled={!file || loading}
              className="w-full px-8 py-3 bg-gray-900 text-white rounded-xl font-bold shadow-md hover:bg-black disabled:opacity-50 transition"
            >
              {loading ? '⚡ Processing...' : 'Start Upload'}
            </button>
         </div>
      </div>

      {result && (
        <div className={`p-8 rounded-3xl border-2 ${result.status === 'success' ? 'bg-emerald-50 border-emerald-100' : 'bg-red-50 border-red-100'} animate-in fade-in zoom-in duration-500`}>
          <h4 className={`text-lg font-bold mb-2 ${result.status === 'success' ? 'text-emerald-800' : 'text-red-800'}`}>
            {result.status === 'success' ? '🎊 Import Completed!' : '❌ Import Failed'}
          </h4>
          <p className="text-sm font-medium opacity-80">
            {result.rows_imported} rows successfully added to the database.
          </p>
        </div>
      )}
    </div>
  );
};

export default DataImporter;
