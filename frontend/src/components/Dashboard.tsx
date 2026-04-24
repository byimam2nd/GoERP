import React, { useEffect, useState } from 'react';

const NumberCard = ({ title, doctype, type, field, period, tenantID, token }: any) => {
  const [value, setValue] = useState<number | null>(null);

  useEffect(() => {
    fetch(`/api/v1/stats/${doctype}?type=${type}&field=${field}&period=${period}`, {
      headers: { 'X-GoERP-Tenant': tenantID, 'Authorization': `Bearer ${token}` }
    })
      .then(res => res.json())
      .then(data => setValue(data.value));
  }, [doctype, type, field, period, tenantID, token]);

  return (
    <div className="bg-white p-6 rounded-3xl border border-gray-100 shadow-sm hover:shadow-xl transition-all group">
      <h3 className="text-[10px] font-black text-gray-400 uppercase tracking-widest mb-1 group-hover:text-indigo-600 transition">{title}</h3>
      <div className="text-3xl font-black text-gray-900 tracking-tighter">
        {value === null ? '...' : (type === 'sum' ? `$${value.toLocaleString()}` : value.toLocaleString())}
      </div>
      <div className="mt-2 text-[10px] font-bold text-gray-300 uppercase italic">
        {period || 'All Time'}
      </div>
    </div>
  );
};

const Dashboard: React.FC<any> = ({ tenantID, token, onNavigate }) => {
  return (
    <div className="max-w-7xl mx-auto p-10 space-y-12">
      <header className="flex justify-between items-center">
        <div>
          <h1 className="text-4xl font-black text-gray-900 tracking-tighter">Command Center</h1>
          <p className="text-gray-400 font-medium">Real-time business intelligence engine.</p>
        </div>
        <div className="flex space-x-2">
           <button onClick={() => onNavigate('DocTypeBuilder')} className="px-4 py-2 bg-black text-white rounded-xl font-bold text-xs uppercase tracking-widest shadow-xl active:scale-95 transition">Architect</button>
        </div>
      </header>

      {/* 1. Key Metrics */}
      <section>
        <h2 className="text-xs font-black text-indigo-600 uppercase tracking-[0.3em] mb-6 flex items-center">
          <span className="w-8 h-[2px] bg-indigo-100 mr-3"></span>
          Key Performance Indicators
        </h2>
        <div className="grid grid-cols-1 md:grid-cols-4 gap-6">
          <NumberCard title="Total Sales" doctype="SalesInvoice" type="sum" field="total" period="month" tenantID={tenantID} token={token} />
          <NumberCard title="Open Orders" doctype="SalesOrder" type="count" period="today" tenantID={tenantID} token={token} />
          <NumberCard title="Stock Value" doctype="Item" type="sum" field="valuation_rate" tenantID={tenantID} token={token} />
          <NumberCard title="New Customers" doctype="Customer" type="count" period="week" tenantID={tenantID} token={token} />
        </div>
      </section>

      {/* 2. Shortcuts & Navigation */}
      <div className="grid grid-cols-1 md:grid-cols-3 gap-10">
        <div className="md:col-span-2 space-y-6">
           <h2 className="text-xs font-black text-indigo-600 uppercase tracking-[0.3em] flex items-center">
            <span className="w-8 h-[2px] bg-indigo-100 mr-3"></span>
            Rapid Access
           </h2>
           <div className="grid grid-cols-2 md:grid-cols-4 gap-4">
              {['SalesInvoice', 'PurchaseOrder', 'Item', 'Customer', 'Account', 'Employee', 'WorkOrder', 'JournalEntry'].map(dt => (
                <button 
                  key={dt}
                  onClick={() => onNavigate('DynamicList', { doctype: dt })}
                  className="p-6 bg-gray-50 rounded-2xl border border-gray-100 text-left hover:bg-white hover:border-indigo-200 hover:shadow-lg transition group"
                >
                  <div className="w-10 h-10 bg-white rounded-lg flex items-center justify-center mb-4 shadow-sm group-hover:bg-indigo-600 group-hover:text-white transition">
                    <span className="font-black text-xs">{dt.substring(0, 2).toUpperCase()}</span>
                  </div>
                  <span className="text-xs font-black text-gray-700 uppercase tracking-tighter">{dt.replace(/([A-Z])/g, ' $1').trim()}</span>
                </button>
              ))}
           </div>
        </div>

        <div className="space-y-6">
           <h2 className="text-xs font-black text-indigo-600 uppercase tracking-[0.3em] flex items-center">
            <span className="w-8 h-[2px] bg-indigo-100 mr-3"></span>
            System Health
           </h2>
           <div className="bg-white rounded-3xl border border-gray-100 p-8 space-y-6 shadow-sm">
              <div className="flex justify-between items-center">
                <span className="text-xs font-bold text-gray-400 uppercase">Database</span>
                <span className="w-2 h-2 bg-green-500 rounded-full animate-pulse"></span>
              </div>
              <div className="flex justify-between items-center">
                <span className="text-xs font-bold text-gray-400 uppercase">Redis Engine</span>
                <span className="w-2 h-2 bg-green-500 rounded-full animate-pulse"></span>
              </div>
              <div className="flex justify-between items-center border-t pt-4">
                <span className="text-xs font-bold text-gray-400 uppercase">Active Tenant</span>
                <span className="text-xs font-black text-indigo-600 uppercase tracking-widest">{tenantID}</span>
              </div>
           </div>
        </div>
      </div>
    </div>
  );
};

export default Dashboard;
