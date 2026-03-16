import React, { useState, useEffect } from 'react';
import { Layout, Input, Button, Empty, message, Card } from 'antd';
import { SendOutlined } from '@ant-design/icons';
import { messageApi, MessageInfo } from '../api/message';
import { useAuthStore } from '../store/authStore';

const { Content } = Layout;

const MessagePage: React.FC = () => {
  const { userInfo } = useAuthStore();
  const [messages, setMessages] = useState<MessageInfo[]>([]);
  const [inputText, setInputText] = useState('');
  const [loading, setLoading] = useState(false);

  useEffect(() => {
    loadMessages();
  }, []);

  const loadMessages = async () => {
    try {
      const result = await messageApi.getMessageList({
        rider_id: Number(userInfo?.userId || 0),
        page: 1,
        page_size: 20
      });
      setMessages(result.messages || []);
    } catch (error) {
      console.error('Load messages error:', error);
    }
  };

  const handleSend = async () => {
    if (!inputText.trim()) return;

    setLoading(true);
    try {
      await messageApi.sendMessage({
        sender_id: Number(userInfo?.userId || 0),
        receiver_id: 0, // 系统消息
        content: inputText.trim()
      });
      setInputText('');
      loadMessages();
      message.success('发送成功');
    } catch (error) {
      message.error('发送失败');
    } finally {
      setLoading(false);
    }
  };

  return (
    <Layout style={{ height: 'calc(100vh - 64px)' }}>
      <Content style={{ display: 'flex', flexDirection: 'column', padding: 24 }}>
        <Card title="消息中心" style={{ marginBottom: 16 }}>
          <p>系统通知和消息</p>
        </Card>
        
        <div style={{ flex: 1, overflowY: 'auto', background: '#f5f5f5', padding: 16, borderRadius: 8 }}>
          {messages.length === 0 ? (
            <Empty description="暂无消息" />
          ) : (
            messages.map((msg) => (
              <div
                key={msg.id}
                style={{
                  display: 'flex',
                  justifyContent: msg.sender_id === Number(userInfo?.userId) ? 'flex-end' : 'flex-start',
                  marginBottom: 12
                }}
              >
                <div
                  style={{
                    maxWidth: '60%',
                    padding: '8px 12px',
                    borderRadius: 8,
                    backgroundColor: msg.sender_id === Number(userInfo?.userId) ? '#FFD700' : '#fff',
                    boxShadow: '0 1px 2px rgba(0,0,0,0.1)'
                  }}
                >
                  {msg.content}
                </div>
              </div>
            ))
          )}
        </div>
        
        <div style={{ display: 'flex', gap: 8, marginTop: 16 }}>
          <Input
            value={inputText}
            onChange={(e) => setInputText(e.target.value)}
            onPressEnter={handleSend}
            placeholder="输入消息..."
          />
          <Button type="primary" icon={<SendOutlined />} onClick={handleSend} loading={loading}>
            发送
          </Button>
        </div>
      </Content>
    </Layout>
  );
};

export default MessagePage;
