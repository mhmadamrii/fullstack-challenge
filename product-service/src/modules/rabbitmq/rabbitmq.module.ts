import { Module } from '@nestjs/common';
import { RabbitMQProvider } from './rabbitmq.provider';

@Module({
  providers: [RabbitMQProvider],
  exports: ['RABBITMQ_CHANNEL'],
})
export class RabbitMQModule {}
