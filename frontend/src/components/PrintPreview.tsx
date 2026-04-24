import React from 'react';

const PrintPreview: React.FC<{ doctype: string, name: string, tenantID: string, token: string }> = ({ doctype, name, tenantID, token }) => {
  const printUrl = `/api/v1/resource/${doctype}/${name}/print?token=${token}`; // For simplified preview

  const handlePrint = () => {
    const frame = document.getElementById('print-frame') as HTMLIFrameElement;
    if (frame && frame.contentWindow) {
      frame.contentWindow.print();
    }
  };

  return (
    <div className="bg-gray-800 min-h-screen p-8 flex flex-col items-center">
      <header className="max-w-4xl w-full flex justify-between items-center mb-6">
        <h1 className="text-white font-black text-xl tracking-tight">Print Preview</h1>
        <div className="space-x-4">
           <button className="px-6 py-2 bg-gray-700 text-gray-300 rounded-lg font-bold">Settings</button>
           <button 
             onClick={handlePrint}
             className="px-6 py-2 bg-blue-600 text-white rounded-lg font-bold shadow-lg hover:bg-blue-700 transition"
           >
             🖨️ Print Now
           </button>
        </div>
      </header>

      <div className="max-w-4xl w-full bg-white shadow-2xl rounded-sm overflow-hidden">
        <iframe 
          id="print-frame"
          src={printUrl} 
          className="w-full h-[1100px] border-none"
          title="Print Preview"
        />
      </div>
    </div>
  );
};

export default PrintPreview;
