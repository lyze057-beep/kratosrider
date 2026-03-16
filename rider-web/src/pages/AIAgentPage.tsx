import React, { useState, useEffect, useRef } from 'react';
import { Layout, Input, Button, List, Avatar, message, Spin, Tag } from 'antd';
import { SendOutlined, UserOutlined, RobotOutlined } from '@ant-design/icons';
import { aiAgentApi, AIAgentChatMessage, MESSAGE_TYPE } from '../api/aiAgent';
import { useAuthStore } from '../store/authStore';

const { Content } = Layout;
const { TextArea } = Input;

const AIAgentPage: React.FC = () => {
  const { userInfo } = useAuthStore();
  const [messages, setMessages] = useState<AIAgentChatMessage[]>([]);
  const [inputText, setInputText] = useState('');
  const [loading, setLoading] = useState(false);
  const [historyLoading, setHistoryLoading] = useState(true);
  const messagesEndRef = useRef<HTMLDivElement>(null);

  const scrollToBottom = () => {
    messagesEndRef.current?.scrollIntoView({ behavior: 'smooth' });
  };

  useEffect(() => {
    scrollToBottom();
  }, [messages]);

  useEffect(() => {
    loadChatHistory();
  }, []);

  const loadChatHistory = async () => {
    try {
      const result = await aiAgentApi.getChatHistory(Number(userInfo?.userId || 0), 20);
      setMessages(result.messages || []);
    } catch (error) {
      console.error('Load chat history error:', error);
    } finally {
      setHistoryLoading(false);
    }
  };

  const handleSend = async () => {
    if (!inputText.trim() || loading) return;

    const content = inputText.trim();
    setInputText('');
    setLoading(true);

    const userMessage: AIAgentChatMessage = {
      id: Date.now(),
      rider_id: Number(userInfo?.userId || 0),
      content,
      message_type: MESSAGE_TYPE.USER,
      content_type: 0,
      created_at: new Date().toISOString(),
    };

    setMessages((prev) => [...prev, userMessage]);

    try {
      const result = await aiAgentApi.sendMessage(
        Number(userInfo?.userId || 0),
        content
      );

      console.log('AI response:', result);

      // 支持驼峰命名和蛇形命名
      const aiResponse = result.ai_response || result.aiResponse;
      
      if (result.success && aiResponse) {
        const aiMessage: AIAgentChatMessage = {
          id: Date.now() + 1,
          rider_id: Number(userInfo?.userId || 0),
          content: aiResponse.content || aiResponse.Content || '',
          message_type: MESSAGE_TYPE.AI,
          content_type: 0,
          created_at: aiResponse.created_at || aiResponse.CreatedAt || new Date().toISOString(),
        };
        setMessages((prev) => [...prev, aiMessage]);
      } else if (!result.success) {
        message.error(result.message || 'AI回复失败');
      }
    } catch (error: any) {
      console.error('Send message error:', error);
      message.error(error.response?.data?.message || '发送失败');
    } finally {
      setLoading(false);
    }
  };

  const renderMessage = (item: AIAgentChatMessage) => {
    const isUser = item.message_type === MESSAGE_TYPE.USER;
    return (
      <List.Item style={{ border: 'none', padding: '8px 0' }}>
        <div style={{ 
          display: 'flex', 
          justifyContent: isUser ? 'flex-end' : 'flex-start',
          width: '100%'
        }}>
          {!isUser && (
            <Avatar icon={<RobotOutlined />} style={{ backgroundColor: '#FFD700', marginRight: 12 }} />
          )}
          <div style={{ 
            maxWidth: '60%', 
            backgroundColor: isUser ? '#FFD700' : '#f0f0f0',
            padding: '12px 16px',
            borderRadius: isUser ? '16px 16px 4px 16px' : '16px 16px 16px 4px',
            color: isUser ? '#333' : '#333'
          }}>
            {item.content}
          </div>
          {isUser && (
            <Avatar icon={<UserOutlined />} style={{ backgroundColor: '#1890ff', marginLeft: 12 }} />
          )}
        </div>
      </List.Item>
    );
  };

  const suggestedQuestions = [
    '我的收入怎么样？',
    '附近有订单吗？',
    '如何提现？',
    '平台规则是什么？'
  ];

  return (
    <Layout style={{ height: 'calc(100vh - 64px)', background: '#f5f5f5' }}>
      <Content style={{ padding: 24, display: 'flex', flexDirection: 'column' }}>
        <div style={{ flex: 1, overflowY: 'auto', paddingBottom: 16 }}>
          {historyLoading ? (
            <div style={{ textAlign: 'center', padding: '50px' }}>
              <Spin size="large" />
            </div>
          ) : (
            <>
              {messages.length === 0 && (
                <div style={{ textAlign: 'center', padding: '40px 20px' }}>
                  <Avatar size={64} icon={<RobotOutlined />} style={{ backgroundColor: '#FFD700', marginBottom: 16 }} />
                  <h2>美团智能助手</h2>
                  <p style={{ color: '#666', marginBottom: 24 }}>有什么可以帮您的？</p>
                  <div style={{ display: 'flex', flexWrap: 'wrap', gap: 8, justifyContent: 'center' }}>
                    {suggestedQuestions.map((q, index) => (
                      <Tag 
                        key={index} 
                        color="blue" 
                        style={{ fontSize: 14, padding: '8px 16px', cursor: 'pointer' }}
                        onClick={() => { setInputText(q); handleSend(); }}
                      >
                        {q}
                      </Tag>
                    ))}
                  </div>
                </div>
              )}
              <List
                dataSource={messages}
                renderItem={renderMessage}
              />
              <div ref={messagesEndRef} />
            </>
          )}
        </div>
        
        <div style={{ 
          display: 'flex', 
          gap: 12,
          padding: 16,
          background: '#fff',
          borderRadius: 12,
          boxShadow: '0 -2px 8px rgba(0,0,0,0.05)'
        }}>
          <TextArea
            value={inputText}
            onChange={(e) => setInputText(e.target.value)}
            onPressEnter={(e) => {
              if (!e.shiftKey) {
                e.preventDefault();
                handleSend();
              }
            }}
            placeholder="请输入您的问题..."
            autoSize={{ minRows: 1, maxRows: 4 }}
            style={{ flex: 1 }}
          />
          <Button 
            type="primary" 
            icon={<SendOutlined />}
            onClick={handleSend}
            loading={loading}
            disabled={!inputText.trim()}
            style={{ height: 'auto', minWidth: 80 }}
          >
            发送
          </Button>
        </div>
      </Content>
    </Layout>
  );
};

export default AIAgentPage;
