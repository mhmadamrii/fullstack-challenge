import { Injectable, Inject } from '@nestjs/common';
import { CreateProductDto } from './product.dto';
import { PrismaService } from 'src/prisma.service';
import type { RedisClientType } from 'redis';
import * as amqp from 'amqplib';

@Injectable()
export class ProductService {
  constructor(
    private prisma: PrismaService,
    @Inject('REDIS_CLIENT') private redis: RedisClientType,
    @Inject('RABBITMQ_CHANNEL') private readonly channel: amqp.Channel,
  ) {
    this.listenToOrderCreated();
  }

  async listenToOrderCreated() {
    this.channel.consume('order.created', async (msg) => {
      if (msg) {
        const order = JSON.parse(msg.content.toString());
        console.log('Order created:', order);

        const product = await this.prisma.products.findUnique({
          where: { id: order.productId },
        });

        if (product) {
          const newQty = product.qty - 1;
          await this.prisma.products.update({
            where: {
              id: order.productId,
            },
            data: {
              qty: newQty,
            },
          });

          if (newQty === 0) {
            console.log(`Product ${product.name} is out of stock`);
          }
        }

        this.channel.ack(msg);
      }
    });
  }

  async createProduct(dto: CreateProductDto) {
    const newProduct = await this.prisma.products.create({ data: dto });
    await this.redis.del('products');

    const payload = Buffer.from(JSON.stringify(newProduct));
    await this.channel.publish('events', 'product.created', payload);
    return newProduct;
  }

  async getAllProducts() {
    const cachedProducts = await this.redis.get('products');
    if (cachedProducts) {
      console.log('Cached products found ✅');
      return JSON.parse(cachedProducts);
    }

    console.log('Cached products missing 🗑️');
    const products = await this.prisma.products.findMany();
    await this.redis.set('products', JSON.stringify(products));
    return products;
  }

  async getProductById(id: string) {
    const cachedProduct = await this.redis.get(`product_${id}`);
    if (cachedProduct) {
      return JSON.parse(cachedProduct);
    }
    const product = await this.prisma.products.findUnique({ where: { id } });
    await this.redis.set(`product_${id}`, JSON.stringify(product));
    return product;
  }

  async createProductWithoutCache(dto: CreateProductDto) {
    return this.prisma.products.create({ data: dto });
  }

  async getAllProductsWithoutCache() {
    return this.prisma.products.findMany();
  }
}
