#!/usr/bin/env python3
"""
CAS IdP Demo Server
简单的HTTP服务器来运行CAS IdP Demo
"""

import http.server
import socketserver
import webbrowser
import os
import sys
from urllib.parse import urlparse, parse_qs

class CASDemoHandler(http.server.SimpleHTTPRequestHandler):
    def do_GET(self):
        # 处理根路径重定向到demo页面
        if self.path == '/':
            self.path = '/cas-idp-demo.html'
        
        # 处理回调URL
        if self.path.startswith('/callback'):
            self.handle_callback()
            return
            
        # 处理其他请求
        return super().do_GET()
    
    def handle_callback(self):
        """处理CAS回调"""
        parsed_url = urlparse(self.path)
        query_params = parse_qs(parsed_url.query)
        
        # 获取ticket参数
        ticket = query_params.get('ticket', [None])[0]
        
        if ticket:
            # 重定向到demo页面并携带ticket参数
            redirect_url = f'/cas-idp-demo.html?ticket={ticket}'
            self.send_response(302)
            self.send_header('Location', redirect_url)
            self.end_headers()
        else:
            # 没有ticket，直接显示demo页面
            self.path = '/cas-idp-demo.html'
            return super().do_GET()

def main():
    # 设置端口 - 避免与Node.js前端冲突
    PORT = 3001
    
    # 检查端口是否被占用
    try:
        with socketserver.TCPServer(("", PORT), CASDemoHandler) as httpd:
            print(f"🚀 CAS IdP Demo Server starting...")
            print(f"📡 Server running at: http://localhost:{PORT}")
            print(f"🔗 Demo URL: http://localhost:{PORT}/cas-idp-demo.html")
            print(f"📋 Callback URL: http://localhost:{PORT}/callback")
            print(f"⚠️  Note: Using port {PORT} to avoid conflict with Node.js frontend on port 3000")
            print(f"⏹️  Press Ctrl+C to stop the server")
            print("-" * 50)
            
            # 自动打开浏览器
            try:
                webbrowser.open(f'http://localhost:{PORT}/cas-idp-demo.html')
                print("🌐 Browser opened automatically")
            except:
                print("⚠️  Could not open browser automatically")
            
            print("-" * 50)
            
            # 启动服务器
            httpd.serve_forever()
            
    except OSError as e:
        if e.errno == 48:  # Address already in use
            print(f"❌ Port {PORT} is already in use. Please stop the service using this port or change the port number.")
            sys.exit(1)
        else:
            print(f"❌ Error starting server: {e}")
            sys.exit(1)
    except KeyboardInterrupt:
        print("\n🛑 Server stopped by user")
        sys.exit(0)

if __name__ == "__main__":
    main()
