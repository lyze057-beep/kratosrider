import React, { useState, useEffect, useRef } from 'react';
import {
  View,
  Text,
  FlatList,
  TextInput,
  TouchableOpacity,
  StyleSheet,
  KeyboardAvoidingView,
  Platform,
  Alert,
} from 'react-native';
import { aiAgentApi, AIAgentChatMessage, MESSAGE_TYPE } from '../api/aiAgent';
import { useAuthStore } from '../store/authStore';

export const AIAgentScreen: React.FC = () => {
  const { userInfo } = useAuthStore();
  const [messages, setMessages] = useState<AIAgentChatMessage[]>([]);
  const [inputText, setInputText] = useState('');
  const [loading, setLoading] = useState(false);
  const [suggestedQuestions, setSuggestedQuestions] = useState<string[]>([]);
  const flatListRef = useRef<FlatList>(null);

  useEffect(() => {
    loadChatHistory();
    loadSuggestedQuestions();
  }, []);

  const loadChatHistory = async () => {
    try {
      const result = await aiAgentApi.getChatHistory({
        rider_id: Number(userInfo?.user_id || 0),
        limit: 20,
      });
      setMessages(result.messages || []);
    } catch (error) {
      console.error('Load chat history error:', error);
    }
  };

  const loadSuggestedQuestions = async () => {
    try {
      const result = await aiAgentApi.getSuggestedQuestions({
        rider_id: Number(userInfo?.user_id || 0),
      });
      setSuggestedQuestions(result.questions || []);
    } catch (error) {
      console.error('Load suggested questions error:', error);
    }
  };

  const handleSend = async (text?: string) => {
    const content = text || inputText.trim();
    if (!content) return;

    setInputText('');
    setLoading(true);

    const userMessage: AIAgentChatMessage = {
      id: Date.now(),
      rider_id: Number(userInfo?.user_id || 0),
      content,
      message_type: MESSAGE_TYPE.USER,
      content_type: 0,
      created_at: new Date().toISOString(),
    };

    setMessages((prev) => [...prev, userMessage]);

    try {
      const result = await aiAgentApi.sendMessage({
        rider_id: Number(userInfo?.user_id || 0),
        content,
        message_type: 0,
      });

      if (result.success && result.ai_response) {
        const aiMessage: AIAgentChatMessage = {
          id: Date.now() + 1,
          rider_id: Number(userInfo?.user_id || 0),
          content: result.ai_response.content,
          message_type: MESSAGE_TYPE.AI,
          content_type: 0,
          created_at: result.ai_response.created_at,
        };
        setMessages((prev) => [...prev, aiMessage]);
      }
    } catch (error: any) {
      Alert.alert('错误', error.response?.data?.message || '发送失败');
    } finally {
      setLoading(false);
    }
  };

  const renderMessage = ({ item }: { item: AIAgentChatMessage }) => {
    const isUser = item.message_type === MESSAGE_TYPE.USER;
    return (
      <View
        style={[
          styles.messageContainer,
          isUser ? styles.userMessageContainer : styles.aiMessageContainer,
        ]}
      >
        {!isUser && (
          <View style={styles.aiAvatar}>
            <Text style={styles.aiAvatarText}>AI</Text>
          </View>
        )}
        <View
          style={[
            styles.messageBubble,
            isUser ? styles.userBubble : styles.aiBubble,
          ]}
        >
          <Text style={[styles.messageText, isUser && styles.userMessageText]}>
            {item.content}
          </Text>
        </View>
      </View>
    );
  };

  return (
    <KeyboardAvoidingView
      style={styles.container}
      behavior={Platform.OS === 'ios' ? 'padding' : undefined}
      keyboardVerticalOffset={90}
    >
      <FlatList
        ref={flatListRef}
        data={messages}
        keyExtractor={(item) => item.id.toString()}
        renderItem={renderMessage}
        contentContainerStyle={styles.messageList}
        onContentSizeChange={() => flatListRef.current?.scrollToEnd()}
        ListHeaderComponent={
          suggestedQuestions.length > 0 && messages.length === 0 ? (
            <View style={styles.suggestedContainer}>
              <Text style={styles.suggestedTitle}>快捷问题</Text>
              {suggestedQuestions.map((question, index) => (
                <TouchableOpacity
                  key={index}
                  style={styles.suggestedBtn}
                  onPress={() => handleSend(question)}
                >
                  <Text style={styles.suggestedBtnText}>{question}</Text>
                </TouchableOpacity>
              ))}
            </View>
          ) : null
        }
      />

      <View style={styles.inputContainer}>
        <TextInput
          style={styles.input}
          placeholder="请输入您的问题..."
          value={inputText}
          onChangeText={setInputText}
          multiline
          maxLength={500}
        />
        <TouchableOpacity
          style={[styles.sendBtn, loading && styles.sendBtnDisabled]}
          onPress={() => handleSend()}
          disabled={loading || !inputText.trim()}
        >
          <Text style={styles.sendBtnText}>{loading ? '...' : '发送'}</Text>
        </TouchableOpacity>
      </View>
    </KeyboardAvoidingView>
  );
};

const styles = StyleSheet.create({
  container: {
    flex: 1,
    backgroundColor: '#f5f5f5',
  },
  messageList: {
    padding: 16,
    paddingBottom: 8,
  },
  messageContainer: {
    flexDirection: 'row',
    marginBottom: 16,
  },
  userMessageContainer: {
    justifyContent: 'flex-end',
  },
  aiMessageContainer: {
    justifyContent: 'flex-start',
  },
  aiAvatar: {
    width: 36,
    height: 36,
    borderRadius: 18,
    backgroundColor: '#FFD700',
    justifyContent: 'center',
    alignItems: 'center',
    marginRight: 8,
  },
  aiAvatarText: {
    fontSize: 14,
    fontWeight: 'bold',
    color: '#333',
  },
  messageBubble: {
    maxWidth: '75%',
    padding: 12,
    borderRadius: 16,
  },
  userBubble: {
    backgroundColor: '#FFD700',
    borderBottomRightRadius: 4,
  },
  aiBubble: {
    backgroundColor: '#fff',
    borderBottomLeftRadius: 4,
  },
  messageText: {
    fontSize: 15,
    color: '#333',
    lineHeight: 22,
  },
  userMessageText: {
    color: '#333',
  },
  suggestedContainer: {
    marginBottom: 16,
  },
  suggestedTitle: {
    fontSize: 14,
    color: '#666',
    marginBottom: 12,
  },
  suggestedBtn: {
    backgroundColor: '#fff',
    paddingVertical: 12,
    paddingHorizontal: 16,
    borderRadius: 8,
    marginBottom: 8,
    borderWidth: 1,
    borderColor: '#eee',
  },
  suggestedBtnText: {
    fontSize: 14,
    color: '#333',
  },
  inputContainer: {
    flexDirection: 'row',
    alignItems: 'flex-end',
    backgroundColor: '#fff',
    padding: 12,
    borderTopWidth: 1,
    borderTopColor: '#eee',
  },
  input: {
    flex: 1,
    minHeight: 40,
    maxHeight: 100,
    borderWidth: 1,
    borderColor: '#ddd',
    borderRadius: 20,
    paddingHorizontal: 16,
    paddingVertical: 10,
    fontSize: 15,
    marginRight: 12,
  },
  sendBtn: {
    width: 60,
    height: 40,
    backgroundColor: '#FFD700',
    borderRadius: 20,
    justifyContent: 'center',
    alignItems: 'center',
  },
  sendBtnDisabled: {
    backgroundColor: '#ccc',
  },
  sendBtnText: {
    fontSize: 15,
    fontWeight: 'bold',
    color: '#333',
  },
});
